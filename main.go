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
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
			log.FieldKeyTime:  "@timestamp",
		},
	})
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
	f.AddParser(reflect.TypeOf(TagsOrigin{}), &TagsOrigin{})
	f.AddParser(reflect.TypeOf([]ProxyRoute{}), &ProxyRoutesValue{})

	s := staert.NewStaert(webServerCmd)
	s.AddSource(f)
	if _, err := s.LoadConfig(); err != nil {
		log.Error(fmt.Sprintf("Error running webserver: %s", err))
		os.Exit(-1)
	}

	if err := s.Run(); err != nil {
		log.Error(fmt.Sprintf("Error running webserver: %s", err))
		os.Exit(-1)
	}
	os.Exit(0)
}

func run(webServerConfiguration *WebServerConfiguration) {
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{
				"error_code":    42,
				"error_message": r,
			}).Error(fmt.Sprintf("Failed to boot: %s", r))
		}
	}()
	s := NewWebServer(webServerConfiguration)
	s.Init()
	s.Run()
}
