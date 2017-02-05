package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"os"
	"reflect"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	webServerConfiguration := NewWebServerConfiguration()
	webServerDefaultPointers := NewWebServerDefaultPointers()
	webServerCmd := &flaeg.Command{
		Name:                  "WebServer",
		Description:           `WebServer`,
		Config:                webServerConfiguration,
		DefaultPointersConfig: webServerDefaultPointers,
		Run: func() error {
			run(webServerConfiguration)
			return nil
		},
	}

	f := flaeg.New(webServerCmd, os.Args[1:])
	f.AddParser(reflect.TypeOf(EntryPoint{}), &EntryPoint{})
	f.AddParser(reflect.TypeOf(ConsulConf{}), &ConsulConf{})
	f.AddParser(reflect.TypeOf(ConsulTags{}), &ConsulTags{})

	s := staert.NewStaert(webServerCmd)
	s.AddSource(f)
	if _, err := s.LoadConfig(); err != nil {
		log.Error("Error running webserver: %s\n", err)
		os.Exit(-1)
	}

	fmt.Println(webServerConfiguration.Consul)
	log.Info("Listening to {}:{}", webServerConfiguration.EntryPoint.Address, webServerConfiguration.EntryPoint.Port)
	if err := s.Run(); err != nil {
		log.Error("Error running webserver: %s\n", err)
		os.Exit(-1)
	}
	os.Exit(0)
}

func run(webServerConfiguration *WebServerConfiguration) {
	s := NewWebServer(webServerConfiguration)
	s.Init()
	s.Run()
}
