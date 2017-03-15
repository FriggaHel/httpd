package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strconv"
	"strings"
)

type WebServerConfiguration struct {
	GlobalConfiguration `mapstructure:",squash"`
	ConfigFile          string `short:"c" description:"Configuration file to use (TOML)"`
}

type GlobalConfiguration struct {
	Debug       bool         `description:"Enable Debug"`
	AngularMode bool         `description:"AngularMode Debug"`
	LogLevel    string       `description:"Log level"`
	EntryPoint  *EntryPoint  `description:"EntryPoint"`
	RootFolder  string       `description:"RootFolder"`
	ServiceName string       `description:"ServiceName"`
	Consul      *ConsulConf  `description:"All config around consul"`
	TagsOrigin  *TagsOrigin  `description:"Configuration of origin of tags (overrides commandlinet tags)"`
	ProxyRoutes []ProxyRoute `description:"Path to proxify (ex: /api/)"`
	PreInitCmd  []PreInitCmd `description:"PreInit commands to run before serving"`
}

// PreInitCmds
type PreInitCmdsValue []PreInitCmd

func (p *PreInitCmdsValue) Get() interface{} {
	return p
}

func (p *PreInitCmdsValue) Set(s string) error {
	*p = append(*p, PreInitCmd{
		Command: s,
	})
	return nil
}

func (p *PreInitCmdsValue) SetValue(val interface{}) {
	*p = PreInitCmdsValue(val.([]PreInitCmd))
}

func (p *PreInitCmdsValue) String() string {
	return fmt.Sprintf("%+v", *p)
}

// PreInitCmd
type PreInitCmd struct {
	Command string
}

func (p *PreInitCmd) Get() interface{} {
	return p
}

func (p *PreInitCmd) Set(s string) error {
	p.Command = s
	return nil
}

func (p *PreInitCmd) String() string {
	return p.Command
}

// Proxy Routes
type ProxyRoute struct {
	Path       string `description:"Path to proxify"`
	Scheme     string `description:"Scheme for backend"`
	Host       string `description:"Backend Hostname"`
	StripPath  bool   `description:"Strip Path before forwarding to backend"`
	PrefixPath string `description:"Add prefix to backend URI"`
}

func (pr *ProxyRoute) Get() interface{} {
	return pr
}

func (pr *ProxyRoute) Set(s string) error {
	segs := strings.Split(s, "|")
	if len(segs) != 5 {
		log.Warning(fmt.Sprintf("Invalid format of ProxyRoute: Expected 5 chunks, got %d", len(segs)))
		panic("Invalid format of ProxyRoute")
	}
	stripPath, err := strconv.ParseBool(segs[3])
	if err != nil {
		log.Warning(fmt.Sprintf("[proxy] %s: Invalid StripPath value, defaulting to false", segs[0]))
	}
	pr.Path = segs[0]
	pr.Scheme = segs[1]
	pr.Host = segs[2]
	pr.StripPath = stripPath
	pr.PrefixPath = segs[4]
	return nil
}

func (to *ProxyRoute) String() string {
	return fmt.Sprintf("%+v", *to)
}

// Proxy Routes
type ProxyRoutesValue []ProxyRoute

func (pr *ProxyRoutesValue) Get() interface{} {
	return pr
}

func (pr *ProxyRoutesValue) Set(s string) error {
	segs := strings.Split(s, "|")
	if len(segs) != 5 {
		log.Warning(fmt.Sprintf("Invalid format of ProxyRoute: Expected 5 chunks, got %d", len(segs)))
		panic("Invalid format of ProxyRoute")
	}
	stripPath, err := strconv.ParseBool(segs[3])
	if err != nil {
		log.Warning(fmt.Sprintf("[proxy] %s: Invalid StripPath value, defaulting to false", segs[0]))
	}
	r := ProxyRoute{
		Path:       segs[0],
		Scheme:     segs[1],
		Host:       segs[2],
		StripPath:  stripPath,
		PrefixPath: segs[4],
	}
	*pr = append(*pr, r)
	return nil
}

func (to *ProxyRoutesValue) String() string {
	return fmt.Sprintf("%+v", *to)
}

func (to *ProxyRoutesValue) SetValue(val interface{}) {
	*to = ProxyRoutesValue(val.([]ProxyRoute))
}

// Tags Origin
type TagsOrigin struct {
	Enabled     bool   `description:"Enable fetching tags across another service"`
	ServiceName string `description:"Service to contact"`
	Path        string `description:"File to fetch"`
	IsFatal     bool   `description:"Abort if unable to fetch tags"`
}

func (to *TagsOrigin) Get() interface{} {
	return to
}

func (to *TagsOrigin) Set(s string) error {
	st := strings.Split(s, ":")
	enable, err := strconv.ParseBool(st[0])
	if err != nil {
		to.Enabled = enable
	}
	to.ServiceName = st[1]
	to.Path = st[2]
	fatal, err := strconv.ParseBool(st[3])
	if err != nil {
		to.IsFatal = fatal
	}
	return nil
}

func (to *TagsOrigin) String() string {
	return fmt.Sprintf("%+v", *to)
}

func (to *TagsOrigin) SetValue(val interface{}) {
	*to = TagsOrigin(val.(TagsOrigin))
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
	return fmt.Sprintf("%+v", *t)
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

// Tags Fetched
type FetchedTags struct {
	Tags ConsulTags `yaml:"tags"`
}

func NewWebServerConfiguration() *WebServerConfiguration {
	return &WebServerConfiguration{
		GlobalConfiguration: GlobalConfiguration{
			Debug:       true,
			AngularMode: true,
			LogLevel:    "DEBUG",
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
			TagsOrigin: &TagsOrigin{
				Enabled:     false,
				ServiceName: "config",
				Path:        "/unkown/config.yml",
				IsFatal:     true,
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

	var tags = TagsOrigin{
		Enabled:     false,
		ServiceName: "config",
		Path:        "/unkown/config.yml",
		IsFatal:     true,
	}

	defaultConfiguration := GlobalConfiguration{
		AngularMode: true,
		EntryPoint:  &entryPoint,
		RootFolder:  "/var/www/html",
		ServiceName: "unknown",
		Consul:      &consul,
		TagsOrigin:  &tags,
	}

	return &WebServerConfiguration{
		GlobalConfiguration: defaultConfiguration,
	}
}
