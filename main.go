package main

import (
	"./player"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
)

const (
	DEF_CONFIG      = "config.json"
	FALLBACK_CONFIG = "sample-config.json"
)

func doit() error {
	var filename string
	flag.StringVar(&filename, "config", DEF_CONFIG, "config file")
	flag.Parse()

	var data []byte
	var err error

	data, err = ioutil.ReadFile(filename)
	if err != nil && filename == DEF_CONFIG {
		data, err = ioutil.ReadFile(FALLBACK_CONFIG)
	}
	if err != nil {
		return err
	}

	config := player.Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	srv := &player.MediaServer{
		Config: config,
	}

	return srv.Serve()
}

func main() {
	if err := doit(); err != nil {
		fmt.Println("error", err)
	}
}
