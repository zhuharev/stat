package stat

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/clarkduvall/hyperloglog"
	"github.com/jinzhu/now"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/ungerik/go-dry"
	"hash/fnv"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	SITE_BUCKET       = "sites"
	URI_BUCKET        = "uri"
	USER_AGENT_BUCKET = "ua"
	USER_ID_BUCKET    = "uuids"
	SESSION_ID_BUCKET = "session"

	HITS_LEDIS_KEY = "hits"

	BINLOGS_DIR = "binlogs"

	HyperlogPrecision uint8 = 7
)

type Store struct {
	db      *bolt.DB
	ldb     *leveldb.DB
	ledisdb *ledis.DB
	binlog  *Binlog

	DbPath             string
	lastBinlogArchived time.Time
}

func NewStore(DbPath string) (*Store, error) {
	if !dry.FileExists(DbPath) {
		e := os.MkdirAll(DbPath, 0777)
		if e != nil {
			return nil, e
		}
	}
	db, e := bolt.Open(filepath.Join(DbPath, "bolt"), 0600, nil)
	if e != nil {
		return nil, e
	}
	e = db.Update(func(tx *bolt.Tx) (e error) {
		buckets := []string{SITE_BUCKET, URI_BUCKET, USER_AGENT_BUCKET, USER_ID_BUCKET,
			SESSION_ID_BUCKET}
		for _, v := range buckets {
			if e != nil {
				continue
			}
			_, e = tx.CreateBucketIfNotExists([]byte(v))
		}
		return
	})
	if e != nil {
		return nil, e
	}

	ldb, e := leveldb.OpenFile(filepath.Join(DbPath, "level"), nil)
	if e != nil {
		return nil, e
	}

	ledisdb, e := ledis.Open(&config.Config{DataDir: filepath.Join(DbPath, "ledis")})
	if e != nil {
		return nil, e
	}
	ledisdbFirst, e := ledisdb.Select(0)
	if e != nil {
		return nil, e
	}

	e = os.MkdirAll(filepath.Join(DbPath, BINLOGS_DIR), 0777)
	if e != nil {
		return nil, e
	}
	bl, e := NewWriteBinLog(filepath.Join(DbPath, BINLOGS_DIR, "today"))
	if e != nil {
		return nil, e
	}

	s := new(Store)
	s.db = db
	s.ldb = ldb
	s.ledisdb = ledisdbFirst
	s.binlog = bl
	s.DbPath = DbPath

	e = s.ArchiveBinlogIfNeeded()

	return s, e
}

func (s *Store) SiteId(host string) (id int64, e error) {
	id, e = s.GetOrInsert(SITE_BUCKET, host)
	return
}

func (s *Store) UriId(uri string) (int64, error) {
	uri = strings.TrimPrefix(uri, "/")
	if uri == "" {
		return 0, nil
	}
	return s.GetOrInsert(URI_BUCKET, uri)
}

func (s *Store) UserAgentId(uri string) (int64, error) {
	return s.GetOrInsert(USER_AGENT_BUCKET, uri)
}

func (s *Store) UserId(uid string) (int64, error) {
	return s.GetOrInsert(USER_ID_BUCKET, uid)
}

func (s *Store) SessionId(ssid string) (int64, error) {
	return s.GetOrInsert(SESSION_ID_BUCKET, ssid)
}

func (s *Store) Binlog() *Binlog {
	return s.binlog
}

func (s *Store) getHyperLoglog(siteId int64) (hl *hyperloglog.HyperLogLogPlus, e error) {
	var (
		bts   []byte
		isNew bool = true
		ok    bool
	)
	key := itob32(siteId)
	if ok, e = s.ldb.Has(key, nil); ok {
		isNew = false
		bts, e = s.ldb.Get(key, nil)
		if e != nil {
			if leveldb.ErrNotFound == e {
				isNew = true
			} else {
				return
			}
		}
	}

	hl, e = hyperloglog.NewPlus(HyperlogPrecision)
	if e != nil {
		return
	}

	if !isNew {
		e = hl.GobDecode(bts)
		if e != nil {
			return
		}
	}

	return hl, nil
}

func (s *Store) TodayUniqueCount(siteId int64) (cnt int64, e error) {
	var (
		hl *hyperloglog.HyperLogLogPlus
	)
	hl, e = s.getHyperLoglog(siteId)
	if e != nil {
		return
	}
	cnt = int64(hl.Count())
	return
}

func (s *Store) TodayUnique(siteId int64, uid int64) (cnt int64, e error) {
	var (
		hl *hyperloglog.HyperLogLogPlus
	)
	hl, e = s.getHyperLoglog(siteId)
	if e != nil {
		return
	}

	h := fnv.New64a()
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(uid))
	io.WriteString(h, hex.EncodeToString(b))
	hl.Add(h)

	// encode and store hyperloglog
	var bts []byte
	bts, e = hl.GobEncode()
	if e != nil {
		return
	}
	e = s.ldb.Put(itob32(siteId), bts, nil)
	if e != nil {
		return
	}

	cnt = int64(hl.Count())

	return
}

func (s *Store) LastFlush(siteId int64) (time.Time, error) {
	bts, e := s.ledisdb.Get(itob32(siteId))
	if e != nil {
		return time.Time{}, e
	}
	if len(bts) != 4 {
		return time.Now().Truncate(time.Hour * 30), nil
	}
	sesc := int64(btoi32(bts))
	t := time.Unix(sesc, 0)
	return t, nil
}

func (s *Store) NeedFlush(siteId int64) (bool, error) {
	lh, e := s.LastFlush(siteId)
	if e != nil {
		return false, e
	}

	// flush counters if new day
	if lh.Before(now.BeginningOfDay()) {
		return true, nil
	}
	return false, nil
}

func (s *Store) SetFlush(siteId int64) error {
	return s.ledisdb.Set(itob32(siteId), itob32(time.Now().Unix()))
}

func (s *Store) Flush(siteId int64) error {
	_, e := s.ledisdb.ZRem([]byte(HITS_LEDIS_KEY), itob32(siteId))
	if e != nil {
		return e
	}
	e = s.ldb.Delete(itob32(siteId), nil)
	if e != nil {
		return e
	}
	e = s.SetFlush(siteId)
	if e != nil {
		return e
	}
	return nil
}

func (s *Store) IncHit(siteId int64) (int64, error) {
	return s.ledisdb.ZIncrBy([]byte(HITS_LEDIS_KEY), 1, itob32(siteId))
}

func (s *Store) HitCount(siteId int64) (cnt int64, e error) {
	cnt, e = s.ledisdb.ZScore([]byte(HITS_LEDIS_KEY), itob32(siteId))
	if e != nil {
		if e == ledis.ErrScoreMiss {
			return 0, nil
		}
	}
	return
}

func (s *Store) Rank(siteId int64) (rank int64, e error) {
	rank, e = s.ledisdb.ZRevRank([]byte(HITS_LEDIS_KEY), itob32(siteId))
	rank++
	return
}

func (s *Store) RefererId(ref string) (siteId int64, uriId int64, e error) {
	if strings.TrimSpace(ref) == "" {
		return 0, 0, nil
	}
	var u *url.URL
	u, e = url.Parse(ref)
	if e != nil {
		return
	}

	siteId, e = s.SiteId(u.Host)
	if e != nil {
		return
	}

	uriStr := strings.TrimPrefix(u.RequestURI(), "/")
	uriId, e = s.UriId(uriStr)
	if e != nil {
		return
	}
	return
}

func (s *Store) HasSite(host string) (bool, error) {
	return s.Has(SITE_BUCKET, host)
}

func (s *Store) ArchiveBinlog() error {
	path, e := s.getTargetArchiveFName()
	if e != nil {
		return e
	}
	dir := filepath.Dir(path)
	e = os.MkdirAll(dir, 0777)
	if e != nil {
		return e
	}
	return s.Binlog().Archive(path)
}

func (s *Store) ArchiveBinlogIfNeeded() error {
	fname := s.Binlog().f.Name()
	size := dry.FileSize(fname)
	need := false
	// 134MB
	if size > 1<<27 {
		need = true
	}
	if time.Since(s.lastBinlogArchived) > time.Hour*24*30 {
		need = true
	}

	if need {
		s.lastBinlogArchived = time.Now()
		return s.ArchiveBinlog()
	}
	return nil
}

func (s *Store) getTargetArchiveFName() (string, error) {
	tPath := filepath.Join(s.DbPath, BINLOGS_DIR, time.Now().Format("2006/1"))
	if !dry.FileExists(tPath) {
		return filepath.Join(tPath, time.Now().Format("2.bl")), nil
	}
	files, e := dry.ListDirFiles(tPath)
	if e != nil {
		return "", e
	}
	dayStr := time.Now().Format("2")
	return filepath.Join(tPath, fmt.Sprintf("%s.%d.bl", dayStr, len(files))), nil
}

func (s *Store) ArchiveBinlogIfNeededEvery(d time.Duration) {
	tick := time.Tick(d)
	for range tick {
		s.ArchiveBinlogIfNeeded()
	}
}
