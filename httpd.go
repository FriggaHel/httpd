package main

import (
	"fmt"
	"github.com/FriggaHel/httpd/version"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"os"
	"os/exec"
	"reflect"
	"strings"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
			log.FieldKeyTime:  "@timestamp",
		},
	})
	log.Info(fmt.Sprintf("Booting httpd %s (%s) - Date: %s)", version.Version, version.Codename, version.BuildDate))
	webServerConfiguration := NewWebServerConfiguration()
	webServerDefaultPointers := NewWebServerDefaultPointers()
	webServerCmd := &flaeg.Command{
		Name:                  "httpd",
		Description:           `httpd`,
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
	f.AddParser(reflect.TypeOf(ConfigOrigin{}), &ConfigOrigin{})
	f.AddParser(reflect.TypeOf(PreInitCmds{}), &PreInitCmds{})
	f.AddParser(reflect.TypeOf(RouteMappings{}), &RouteMappings{})
	f.AddParser(reflect.TypeOf(ConsulTags{}), &ConsulTags{})

	toml := staert.NewTomlSource("httpd", []string{"/etc/", "."})

	s := staert.NewStaert(webServerCmd)
	s.AddSource(f)
	s.AddSource(toml)
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

	// Load config from ConfigServer
	if webServerConfiguration.ConfigOrigin.Enabled == true {
		cfg := NewConfigFetcher(webServerConfiguration.Consul, webServerConfiguration.ConfigOrigin).Config()
		for k, v := range cfg.Proxies {
			webServerConfiguration.RouteMappings[k] = v
		}
		for _, v := range cfg.Tags {
			webServerConfiguration.ConsulTags = append(webServerConfiguration.ConsulTags, v)
		}
	}

	for _, x := range webServerConfiguration.PreInitCmds {
		log.Info(fmt.Sprintf("[pre-init] Running '%s'", x.Command))
		params := strings.Split(x.Command, " ")
		cmd := exec.Command(params[0])
		cmd.Args = params
		err := cmd.Run()
		if err != nil {
			log.Fatal(fmt.Sprintf("[pre-init] Failure (%s)", err))
			os.Exit(1)
		}
	}
	s := NewWebServer(webServerConfiguration)
	s.Init()
	s.Run()
}
