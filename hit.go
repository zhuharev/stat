package stat

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Hit struct {
	Created   time.Time
	Host      string
	Uri       string
	Referer   string
	Ip        string
	UserAgent string
	Ssid      string
	Uid       string

	siteId int64

	store *Store
}

func NewHit(host, uri, referer, ip, userAgent, ssid, uid string) *Hit {
	h := &Hit{
		Created:   time.Now(),
		Host:      host,
		Uri:       uri,
		Referer:   referer,
		Ip:        ip,
		UserAgent: userAgent,
		Ssid:      ssid,
		Uid:       uid,
	}
	return h
}

var (
	UidCookieName  = "u"
	SsidCookieName = "s"
)

func NewHitFromRequest(req *http.Request) (*Hit, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	ip, _, e := net.SplitHostPort(req.RemoteAddr)
	if e != nil {
		return nil, e
	}

	ssid := "new"
	c, _ := req.Cookie(SsidCookieName)
	if c != nil {
		ssid = c.Value
	}
	uid := "new"
	c, _ = req.Cookie(UidCookieName)
	if c != nil {
		uid = c.Value
	}

	u, ref, e := getUrl(req.URL.RawQuery, req.Referer())
	if e != nil {
		return nil, e
	}

	return NewHit(u.Host, u.RequestURI(), ref, ip, req.UserAgent(), ssid, uid), nil
}

func getUrl(rawquery, referer string) (u *url.URL, ref string, e error) {
	var (
		us string
	)
	arr := strings.Split(rawquery, ";")
	for _, v := range arr {
		switch v[0] {
		case 'r':
			ref = v[1:]
		case 'u':
			us = v[1:]
		}
	}
	if us == "" {
		us = referer
	}
	us, e = url.QueryUnescape(us)
	if e != nil {
		return
	}
	u, e = url.Parse(us)
	return
}

func (h *Hit) SetStore(store *Store) {
	h.store = store
}

func (h *Hit) AsBytes() ([]byte, error) {
	var (
		data []byte
		e    error
	)
	now := itob32(h.Created.Unix())

	uriId, e := h.store.UriId(h.Uri)
	if e != nil {
		return nil, e
	}

	refSite, refUri, e := h.store.RefererId(h.Referer)
	if e != nil {
		return nil, e
	}

	ipBytes := []byte(net.ParseIP(h.Ip).To4())

	uaId, e := h.store.UserAgentId(h.UserAgent)
	if e != nil {
		return nil, e
	}

	ssId, e := h.store.SessionId(h.Ssid)
	if e != nil {
		return nil, e
	}

	uuid, e := h.store.UserId(h.Uid)
	if e != nil {
		return nil, e
	}

	data = append(data, now...)              // 4 byte
	data = append(data, itob32(h.siteId)...) // 4 byte
	data = append(data, itob32(uriId)...)    // 4 byte
	data = append(data, itob32(refSite)...)  // 4 byte
	data = append(data, itob32(refUri)...)   // 4 byte
	data = append(data, ipBytes...)          // 4 byte
	data = append(data, itob32(uaId)...)     // 4 byte
	data = append(data, itob32(ssId)...)     // 4 byte
	data = append(data, itob32(uuid)...)     // 4 bytes
	//                                       // 36 bytes
	return data, nil
}

func (h *Hit) SiteId() (int64, error) {
	var e error
	if h.siteId == 0 {
		h.siteId, e = h.store.SiteId(h.Host)
		if e != nil {
			return 0, e
		}
	}
	return h.siteId, nil
}
