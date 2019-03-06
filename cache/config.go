package main

// Rule cache a URL request or not, and how
type Rule struct {
	Regexp  string `json:"regexp"`
	Enabled bool   `json:"enabled"`
}

// Config of cache extension
type Config struct {
	Rules []*Rule `json:"rules"`
}
