package opt

import (
	"flag"
	"time"
)

var (
	// DefaultExpiration : default kv storage expiration
	DefaultExpiration = 1 * 24 * time.Hour
	// Prefix : default key's prefix for kv storage
	DefaultPrefix = flag.String("default_prefix", "impression::", "default_prefix")
	Prefix        = *DefaultPrefix
)

// Option : option to kv storage configuration
type Option struct {
	Address    string        `json:"address"`
	Password   string        `json:"password"`
	Expiration time.Duration `json:"expiration"`
}
