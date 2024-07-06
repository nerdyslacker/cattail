package types

import (
	"time"
)

type Namespace struct {
	Name  string `json:"name"`
	Peers []Peer `json:"peers"`
}

type Peer struct {
	ID             string    `json:"id"`
	DNSName        string    `json:"dns_name"`
	Name           string    `json:"name"`
	ExitNode       bool      `json:"exit_node"`
	ExitNodeOption bool      `json:"exit_node_option"`
	Online         bool      `json:"online"`
	OS             string    `json:"os"`
	Addrs          []string  `json:"addrs"`
	Routes         []string  `json:"routes"` // primary routes
	IPs            []string  `json:"ips"`
	Created        time.Time `json:"created_at"`
	LastSeen       time.Time `json:"last_seen"`
}

type File struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}
