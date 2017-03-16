package main

import (
	"github.com/hashicorp/consul/api"
)

type WebServerConfiguration struct {
	GlobalConfiguration `mapstructure:",squash"`
	ConfigFile          string `short:"c" description:"Configuration file to use (TOML)"`
}

type GlobalConfiguration struct {
	Debug         bool          `description:"Enable Debug"`
	AngularMode   bool          `description:"AngularMode Debug"`
	LogLevel      string        `description:"Log level"`
	EntryPoint    *EntryPoint   `description:"EntryPoint"`
	RootFolder    string        `description:"RootFolder"`
	ServiceName   string        `description:"ServiceName"`
	Consul        *ConsulConf   `description:"All config around consul"`
	ConfigOrigin  *ConfigOrigin `description:"Origin of conf (tags/proxies)"`
	PreInitCmds   PreInitCmds   `description:"PreInit commands to run before serving"`
	ConsulTags    ConsulTags    `description:"List of consul tags to inject on registering"`
	RouteMappings RouteMappings `description:"Mapping for reversed-proxified routes"`
}

// PreInitCmds
type PreInitCmd struct {
	Command string
}

type PreInitCmds map[string]*PreInitCmd

// ConfigOrigin
type ConfigOrigin struct {
	Enabled     bool   `description:"Enable Fetching"`
	ServiceName string `description:"Service Origin"`
	Path        string `description:"Config file"`
	IsFatal     bool   `description:"Stop boot on failure"`
}

// Consul Config
type ConsulConf struct {
	Register bool   `description:"Enable Consul Connector"`
	Host     string `description:"Consul Host"`
	Port     int    `description:"Consul Port"`
}

// EntryPoint
type EntryPoint struct {
	Address string `description:"Address to bind"`
	Port    int    `description:"Port to bind"`
}

// ConsulTags
type ConsulTags []string

// RouteMappings
type RouteMappings map[string]*RouteMapping
type RouteMapping struct {
	Path       string `description:"Path to map" yaml:"path"`
	Scheme     string `description:"Target Scheme" yaml:"scheme"`
	Host       string `description:"Target Host" yaml:"host"`
	StripPath  bool   `description:"Remove Path before proxify" yaml:"strip_prefix"`
	PrefixPath string `description:"Add prefix before proxify" yaml:"prefix_path"`
}

// Remote Config
type RemoteConfig struct {
	Tags    []string                 `yaml:"tags"`
	Proxies map[string]*RouteMapping `yaml:"proxies"`
}

type ConfigFetcher struct {
	ConsulApi    *api.Client
	ConsulConf   *ConsulConf
	ConfigOrigin *ConfigOrigin
	Catalog      *api.Catalog
}
