// Package source defines vendoring source contracts without importing adapters.
package source

import (
	"context"
)

type Resolved struct {
	URL       string `yaml:"url,omitempty"       json:"url,omitempty"`
	Commit    string `yaml:"commit,omitempty"    json:"commit,omitempty"`
	Version   string `yaml:"version,omitempty"   json:"version,omitempty"`
	Integrity string `yaml:"integrity,omitempty" json:"integrity,omitempty"`
	SHA256    string `yaml:"sha256,omitempty"    json:"sha256,omitempty"`
}

type Result struct {
	Root     string
	Resolved Resolved
}

// Adapter uses locked identities when locked is non-nil. Otherwise it resolves
// the manifest selector. All downloaded files must stay within workspace.
type Adapter interface {
	Fetch(context.Context, Config, *Resolved, string) (Result, error)
}
