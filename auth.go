package eventdb

import (
	"github.com/pelletier/go-toml/v2"
	"os"
)

// Authorizer is a simple authorizer that checks if a user is allowed to access a source.
type Authorizer struct {
	authentication map[string]string
	authorization  map[string]string
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

	return &Authorizer{authorization: a.Authorization, authentication: a.Authentication}, nil
}

// IsAllowed returns true if the user is allowed to access the source.
func (a *Authorizer) IsAllowed(user, source string) bool {
	if u, ok := a.authorization[source]; ok {
		return u == user
	}
	return false
}
