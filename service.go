package stat

import (
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/fatih/color"
	"net/http"
	"time"
)

type Service struct {
	*Store
}

func New(dataPath string) (*Service, error) {
	s := new(Service)
	st, e := NewStore(dataPath)
	if e != nil {
		return nil, e
	}
	s.Store = st
	return s, nil
}

func (s *Service) AddSite(host string) error {
	if ok, e := s.Store.HasSite(host); ok {
		if e != nil {
			return e
		}
		return fmt.Errorf("already added")
	}
	id, e := s.Store.SiteId(host)
	if e != nil {
		return e
	}
	color.Green("added site %s with id %d", host, id)
	return nil
}

var (
	ErrUnknownSite = fmt.Errorf("unknown site")

	UidCookieMaxAge = 60 * 60 * 24 * 366 * 5
)

func (s *Service) HandleHit(w http.ResponseWriter, req *http.Request) error {
	h, e := NewHitFromRequest(req)
	if e != nil {
		return e
	}
	if ok, e := s.Store.HasSite(h.Host); !ok {
		if e != nil {
			return e
		}
		return ErrUnknownSite
	}
	if h.Ssid == "new" {
		h.Ssid = uniuri.New()
		newCookie := &http.Cookie{Name: SsidCookieName, Value: h.Ssid}
		http.SetCookie(w, newCookie)
	}
	if h.Uid == "new" {
		h.Uid = uniuri.New()
		now := time.Now()
		newCookie := &http.Cookie{Name: UidCookieName, Value: h.Uid, MaxAge: UidCookieMaxAge, Expires: now.Add(time.Duration(UidCookieMaxAge))}
		http.SetCookie(w, newCookie)
	}

	h.SetStore(s.Store)
	go s.Save(h)
	return nil
}

func (s *Service) Hit(host string, uri string, referer, ip string, userAgent string, ssid string, uid string) error {
	if s.Store == nil {
		return fmt.Errorf("store not inited")
	}
	if ok, e := s.Store.HasSite(host); !ok {
		if e != nil {
			return e
		}
		return ErrUnknownSite
	}
	h := NewHit(host, uri, referer, ip, userAgent, ssid, uid)
	h.SetStore(s.Store)
	go s.Save(h)
	return nil
}

func (s *Service) Stat(host string) (*Stat, error) {
	if ok, e := s.Store.HasSite(host); !ok {
		if e != nil {
			return nil, e
		}
		return nil, ErrUnknownSite
	}

	siteId, e := s.Store.SiteId(host)
	if e != nil {
		return nil, e
	}

	if ok, e := s.Store.NeedFlush(siteId); ok {
		if e != nil {
			return nil, e
		}
		e = s.Store.Flush(siteId)
		if e != nil {
			return nil, e
		}
	}

	score, e := s.Store.HitCount(siteId)
	if e != nil {
		return nil, e
	}
	rank, e := s.Store.Rank(siteId)
	if e != nil {
		return nil, e
	}
	uniq, e := s.Store.TodayUniqueCount(siteId)
	if e != nil {
		return nil, e
	}
	return &Stat{
		SiteId:    siteId,
		TodayHit:  score,
		Rank:      rank,
		TodayUniq: uniq,
	}, nil
}

func (s *Service) Save(h *Hit) {
	h.SetStore(s.Store)
	bytes, e := h.AsBytes()
	if e != nil {
		color.Red("%s", e)
		// TODO
		_ = e
	}
	e = s.Store.Binlog().Append(bytes)
	if e != nil {
		color.Red("%s", e)
		// TODO
		_ = e
	}

	siteId, e := s.Store.SiteId(h.Host)
	if e != nil {
		color.Red("%s", e)
		// TODO
		_ = e
	}

	if ok, e := s.Store.NeedFlush(siteId); ok {
		if e != nil {
			color.Red("%s", e)
			// TODO
			_ = e
		}
		e = s.Store.Flush(siteId)
		if e != nil {
			color.Red("%s", e)
			// TODO
			_ = e
		}
	}

	_, e = s.Store.IncHit(siteId)
	if e != nil {
		color.Red("%s", e)
		// TODO
		_ = e
	}

	uid, e := s.Store.UserId(h.Uid)
	if e != nil {
		color.Red("%s", e)
		// TODO
		_ = e
	}

	_, e = s.Store.TodayUnique(siteId, uid)
	if e != nil {
		color.Red("%s", e)
		// TODO
		_ = e
	}

}
