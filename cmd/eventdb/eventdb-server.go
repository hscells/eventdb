package main

import (
	"eventdb"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type server struct {
	Host         string   `toml:"host"`
	AllowOrigins []string `toml:"allowed_origins"`
	LogFile      string   `toml:"log_file"`
}

type settings struct {
	AuthFile   string `toml:"auth_file"`
	EventsFile string `toml:"events_file"`
	Server     server `toml:"server"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: eventdb-server <config.toml>")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	var s settings
	err = toml.NewDecoder(f).Decode(&s)
	if err != nil {
		panic(err)

	}

	authorizer, err := eventdb.OpenAuthorizer(s.AuthFile)
	if err != nil {
		panic(err)
	}
	store, err := eventdb.OpenStore(s.EventsFile)
	if err != nil {
		panic(err)
	}

	srv := eventdb.NewServer(authorizer, store, s.Server.AllowOrigins, s.Server.LogFile)
	panic(srv.Serve(s.Server.Host))
}
