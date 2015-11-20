package stat

import (
	"net"
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
}

func NewHit(host, uri, referer, ip, userAgent, ssid, uid string) *Hit {
	return &Hit{
		Created:   time.Now(),
		Host:      host,
		Uri:       uri,
		Referer:   referer,
		Ip:        ip,
		UserAgent: userAgent,
		Ssid:      ssid,
		Uid:       uid,
	}
}

func (h *Hit) AsBytes(store *Store) ([]byte, error) {
	var data []byte
	now := itob32(h.Created.Unix())

	siteId, e := store.SiteId(h.Host)
	if e != nil {
		return nil, e
	}

	uriId, e := store.UriId(h.Uri)
	if e != nil {
		return nil, e
	}

	refSite, refUri, e := store.RefererId(h.Referer)
	if e != nil {
		return nil, e
	}

	ipBytes := []byte(net.ParseIP(h.Ip).To4())

	uaId, e := store.UserAgentId(h.UserAgent)
	if e != nil {
		return nil, e
	}

	ssId, e := store.SessionId(h.Ssid)
	if e != nil {
		return nil, e
	}

	uuid, e := store.UserId(h.Uid)
	if e != nil {
		return nil, e
	}

	data = append(data, now...)             // 4 byte
	data = append(data, itob32(siteId)...)  // 4 byte
	data = append(data, itob32(uriId)...)   // 4 byte
	data = append(data, itob32(refSite)...) // 4 byte
	data = append(data, itob32(refUri)...)  // 4 byte
	data = append(data, ipBytes...)         // 4 byte
	data = append(data, itob32(uaId)...)    // 4 byte
	data = append(data, itob32(ssId)...)    // 4 byte
	data = append(data, itob32(uuid)...)    // 4 bytes
	//                                     // 36 bytes
	return data, nil
}
