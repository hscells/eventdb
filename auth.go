package eventdb

import (
	"github.com/pelletier/go-toml/v2"
	"os"
)

// Authorizer is a simple authorizer that checks if a user is allowed to access a source.
type Authorizer struct {
	authentication map[string]string
}

// OpenAuthorizer opens a new authorizer at the given path.
func OpenAuthorizer(path string) (*Authorizer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	type auth struct {
		Authentication map[string]string `toml:"authentication"`
		Authorization  map[string]string `toml:"authorization"`
	}
	var a auth

	err = toml.NewDecoder(f).Decode(&a)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	return &Authorizer{authentication: a.Authentication}, nil
}
