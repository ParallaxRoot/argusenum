package models

import "net"

type Subdomain struct {
	Name       string   `json:"name"`        // e.g. "api.example.com"
	Domain     string   `json:"domain"`      // e.g. "example.com"
	Source     string   `json:"source"`      // e.g. "crtsh", "bruteforce"
	IPs        []net.IP `json:"ips"`         // resolved IPs (if any)
	Alive      bool     `json:"alive"`       // HTTP reachable
	HTTPStatus int      `json:"http_status"` // last HTTP status
	Tech       []string `json:"tech"`        // tech fingerprint
	Tags       []string `json:"tags"`        // tags contextual (dev, prod, etc)
}