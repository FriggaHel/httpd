package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

// PreInitCmds
func (p *PreInitCmds) Get() interface{} {
	return p
}

func (p *PreInitCmds) Set(s string) error {
	regex := regexp.MustCompile("Name:(?P<Name>\\S*)\\s*Command:(?P<Command>.*)")
	match := regex.FindAllStringSubmatch(s, -1)
	if match == nil {
		return errors.New("Bad PreInitCmds format: " + s)
	}
	matchResult := match[0]
	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			result[name] = matchResult[i]
		}
	}
	c := &PreInitCmd{
		Command: result["Command"],
	}
	(*p)[result["Name"]] = c
	return nil
}

func (p *PreInitCmds) SetValue(val interface{}) {
	*p = PreInitCmds(val.(PreInitCmds))
}

func (p *PreInitCmds) String() string {
	return fmt.Sprintf("%+v", *p)
}

// Origin of configuration
func (to *ConfigOrigin) Get() interface{} {
	return to
}

func (to *ConfigOrigin) Set(s string) error {
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

func (to *ConfigOrigin) String() string {
	return fmt.Sprintf("%+v", *to)
}

func (to *ConfigOrigin) SetValue(val interface{}) {
	*to = ConfigOrigin(val.(ConfigOrigin))
}

// Consul configuration
func (cf *ConsulConf) Get() interface{} {
	return cf
}

func (cf *ConsulConf) Set(s string) error {
	regex := regexp.MustCompile("Host:(?P<Host>\\S*)\\s*Port:(?P<Port>\\S*)\\s*Register:?P<Register>\\S*)")
	match := regex.FindAllStringSubmatch(s, -1)
	if match == nil {
		return errors.New("Bad PreInitCmds format: " + s)
	}
	matchResult := match[0]
	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			result[name] = matchResult[i]
		}
	}
	enable, err := strconv.ParseBool(result["Register"])
	if err != nil {
		cf.Register = enable
	}
	cf.Host = result["Host"]
	port, err := strconv.ParseInt(result["Port"], 10, 32)
	if err != nil {
		cf.Port = int(port)
	}
	return nil
}

func (cf *ConsulConf) String() string {
	return fmt.Sprintf("%+v", *cf)
}

func (cf *ConsulConf) SetValue(val interface{}) {
	*cf = ConsulConf(val.(ConsulConf))
}

// Entrypoint
func (ep *EntryPoint) Get() interface{} {
	return ep
}

func (ep *EntryPoint) Set(s string) error {
	fmt.Println("42")
	regex := regexp.MustCompile("Address:(?P<Address>\\S*)\\s*Port:(?P<Port>\\S*)")
	match := regex.FindAllStringSubmatch(s, -1)
	if match == nil {
		return errors.New("Bad EntryPoint format: " + s)
	}
	matchResult := match[0]
	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			result[name] = matchResult[i]
		}
	}

	ep.Address = result["Address"]
	port, err := strconv.ParseInt(result["Port"], 10, 32)
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

// RouteMapping
func (rm *RouteMappings) Get() interface{} {
	return rm
}

func (rm *RouteMappings) Set(s string) error {
	regex := regexp.MustCompile("Name:(?P<Name>\\S*)\\s*Path:(?P<Path>\\S*)\\s*Scheme:(?P<Scheme>\\S*)\\s*Host:(?P<Host>\\S*)\\s*StripPath:(?P<StripPath>\\S*)\\s*PrefixPath:(?P<PrefixPath>\\S*)")
	match := regex.FindAllStringSubmatch(s, -1)
	if match == nil {
		return errors.New("Bad RouteMappings format: " + s)
	}
	matchResult := match[0]
	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			result[name] = matchResult[i]
		}
	}
	stripPath, err := strconv.ParseBool(result["StripPath"])
	if err != nil {
		log.Warning(fmt.Sprintf("[proxy] %s: Invalid StripPath value, defaulting to false", result["StripPath"]))
	}
	r := &RouteMapping{
		Path:       result["Path"],
		Scheme:     result["Scheme"],
		Host:       result["Host"],
		StripPath:  stripPath,
		PrefixPath: result["PrefixPath"],
	}
	(*rm)[result["Name"]] = r
	return nil
}

func (rm *RouteMappings) String() string {
	return fmt.Sprintf("%+v", *rm)
}

func (rm *RouteMappings) SetValue(val interface{}) {
	*rm = RouteMappings(val.(RouteMappings))
}

// ConsulTags
func (ct *ConsulTags) Get() interface{} {
	return ct
}

func (ct *ConsulTags) Set(s string) error {
	*ct = append(*ct, s)
	return nil
}

func (ct *ConsulTags) String() string {
	return fmt.Sprintf("%+v", *ct)
}

func (ct *ConsulTags) SetValue(val interface{}) {
	*ct = ConsulTags(val.(ConsulTags))
}

// Constructors
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
			},
			ConfigOrigin: &ConfigOrigin{
				Enabled:     false,
				ServiceName: "config",
				Path:        "/unkown/config.yml",
				IsFatal:     true,
			},
			ConsulTags:    []string{},
			RouteMappings: RouteMappings{},
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
	}

	var conf = ConfigOrigin{
		Enabled:     false,
		ServiceName: "config",
		Path:        "/unkown/config.yml",
		IsFatal:     true,
	}

	defaultConfiguration := GlobalConfiguration{
		AngularMode:  true,
		EntryPoint:   &entryPoint,
		RootFolder:   "/var/www/html",
		ServiceName:  "unknown",
		Consul:       &consul,
		ConfigOrigin: &conf,
	}

	return &WebServerConfiguration{
		GlobalConfiguration: defaultConfiguration,
	}
}
