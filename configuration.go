package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type WebServerConfiguration struct {
	GlobalConfiguration `mapstructure:",squash"`
	ConfigFile          string `short:"c" description:"Configuration file to use (TOML)"`
}

type GlobalConfiguration struct {
	Debug       bool        `description:"Enable Debug"`
	LogLevel    string      `description:"Log level"`
	EntryPoint  *EntryPoint `description:"EntryPoint"`
	RootFolder  string      `description:"RootFolder"`
	ServiceName string      `description:"ServiceName"`
	Consul      *ConsulConf `description:"All config around consul"`
}

// Consul Config
type ConsulConf struct {
	Register bool       `description:"Enable Consul Connector"`
	Host     string     `description:"Consul Host"`
	Port     int        `description:"Consul Port"`
	Tags     ConsulTags `description:"Consul Tags"`
}

func (cf *ConsulConf) Get() interface{} {
	return cf
}

func (cf *ConsulConf) Set(s string) error {
	st := strings.Split(s, ":")
	enable, err := strconv.ParseBool(st[0])
	if err != nil {
		cf.Register = enable
	}
	cf.Host = st[1]
	port, err := strconv.ParseInt(st[2], 10, 32)
	if err != nil {
		cf.Port = int(port)
	}
	cf.Tags = strings.Split(st[3], ",")
	return nil
}

func (cf *ConsulConf) String() string {
	return fmt.Sprintf("%+v", *cf)
}

func (cf *ConsulConf) SetValue(val interface{}) {
	*cf = ConsulConf(val.(ConsulConf))
}

// Consul Tags
type ConsulTags []string

func (t *ConsulTags) Set(val string) error {
	tags := strings.Split(val, ",")
	if len(tags) == 0 {
		return errors.New("Bad Tags format: " + val)
	}
	for _, tag := range tags {
		*t = append(*t, tag)
	}
	return nil
}

func (t *ConsulTags) Get() interface{} {
	return ConsulTags(*t)
}

func (t *ConsulTags) SetValue(val interface{}) {
	*t = ConsulTags(val.(ConsulTags))
}

func (t *ConsulTags) String() string {
	return fmt.Sprint("%+v", *t)
}

// EntryPoint
type EntryPoint struct {
	Address string `description:"Address to bind"`
	Port    int    `description:"Port to bind"`
}

func (ep *EntryPoint) Get() interface{} {
	return ep
}

func (ep *EntryPoint) Set(s string) error {
	st := strings.Split(s, ":")
	ep.Address = st[0]
	port, err := strconv.ParseInt(st[1], 10, 32)
	if err != nil {
		ep.Port = int(port)
	}
	return nil
}

func (ep *EntryPoint) String() string {
	return fmt.Sprintf("%+v", *ep)
}

func (ep *EntryPoint) SetValue(val interface{}) {
	*ep = EntryPoint(val.(EntryPoint))
}

func NewWebServerConfiguration() *WebServerConfiguration {
	return &WebServerConfiguration{
		GlobalConfiguration: GlobalConfiguration{
			Debug:    true,
			LogLevel: "DEBUG",
			EntryPoint: &EntryPoint{
				Address: "0.0.0.0",
				Port:    0,
			},
			RootFolder:  "/var/www/html",
			ServiceName: "unknown",
			Consul: &ConsulConf{
				Register: true,
				Host:     "127.0.0.1",
				Port:     8500,
				Tags:     []string{"traefik.enable=false"},
			},
		},
		ConfigFile: "",
	}
}

func NewWebServerDefaultPointers() *WebServerConfiguration {
	var entryPoint = EntryPoint{
		Address: "0.0.0.0",
		Port:    0,
	}

	var consul = ConsulConf{
		Register: true,
		Host:     "127.0.0.1",
		Port:     8500,
		Tags:     []string{"traefik.enable=false"},
	}

	defaultConfiguration := GlobalConfiguration{
		EntryPoint:  &entryPoint,
		RootFolder:  "/var/www/html",
		ServiceName: "unknown",
		Consul:      &consul,
	}

	return &WebServerConfiguration{
		GlobalConfiguration: defaultConfiguration,
	}
}
