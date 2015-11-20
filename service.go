package stat

import (
	"fmt"
)

type Service struct {
	*Store
	bl *Binlog
}

func New() (*Service, error) {
	s := new(Service)
	return s, nil
}

func (s *Service) AddSite(host string) error {
	if ok, _ := s.Store.HasSite(host); ok {
		return fmt.Errorf("already added")
	}
	e := s.AddSite(host)
	if e != nil {
		return e
	}
	return nil
}

func (s *Service) Hit(host string, uri string, referer, ip string, userAgent string, ssid string, uid string) error {
	if ok, _ := s.Store.HasSite(host); !ok {
		return fmt.Errorf("unknown site")
	}
	h := NewHit(host, uri, referer, ip, userAgent, ssid, uid)
	go s.Save(h)
	return nil
}

func (s *Service) Save(h *Hit) {
	bytes, e := h.AsBytes(s.Store)
	e = s.bl.Append(bytes)
	if e != nil {
		// TODO
		_ = e
	}
}
