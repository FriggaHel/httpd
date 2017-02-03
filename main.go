package main

import (
	"fmt"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	fmtlog "log"
	"os"
	"reflect"
)

func main() {
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
	//toml := staert.NewTomlSource("webserver", []string{webServerConfiguration.ConfigFile, "."})

	//s.AddSource(toml)
	fmt.Println(f)
	s.AddSource(f)
	if _, err := s.LoadConfig(); err != nil {
		fmt.Println(err)
		fmt.Println("42")
		os.Exit(-1)
	}

	fmt.Println(webServerConfiguration.Consul)
	if err := s.Run(); err != nil {
		fmtlog.Printf("Error running webserver: %s\n", err)
		os.Exit(-1)
	}
	os.Exit(0)
}

func run(webServerConfiguration *WebServerConfiguration) {
	s := NewWebServer(webServerConfiguration)
	s.Init()
	s.Run()
}
