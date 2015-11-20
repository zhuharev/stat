package stat

import (
	"github.com/boltdb/bolt"
	"net/url"
	"strings"
)

var (
	SITE_BUCKET       = "sites"
	URI_BUCKET        = "uri"
	USER_AGENT_BUCKET = "ua"
	USER_ID_BUCKET    = "uuids"
	SESSION_ID_BUCKET = "session"
)

type Store struct {
	db *bolt.DB
}

func NewStore(boltDbPath string) (*Store, error) {
	db, e := bolt.Open(boltDbPath, 0600, nil)
	if e != nil {
		return nil, e
	}
	s := new(Store)
	s.db = db
	return s, nil
}

func (s *Store) SiteId(host string) (int64, error) {
	return s.GetOrInsert(SITE_BUCKET, host)
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

func (s *Store) RefererId(ref string) (siteId int64, uriId int64, e error) {
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
	return s.Has("sites", host)
}
