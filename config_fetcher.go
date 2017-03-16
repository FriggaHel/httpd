package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/go-yaml/yaml"
	"github.com/hashicorp/consul/api"
	"net/http"
)

func (cf *ConfigFetcher) Config() RemoteConfig {
	var err error = nil

	log.Info(fmt.Sprintf("Loading tags from %s on %s:%d",
		cf.ConfigOrigin.ServiceName,
		cf.ConsulConf.Host,
		cf.ConsulConf.Port))

	// Connect to consul
	cf.ConsulApi, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		panic("Unable to connect to consul")
	}

	// Fetch Service Catalog
	cf.Catalog = cf.ConsulApi.Catalog()
	svcs, _, err := cf.Catalog.Service(cf.ConfigOrigin.ServiceName, "", nil)
	if err != nil {
		if cf.ConfigOrigin.IsFatal == true {
			panic("Unable to get catalog")
		} else {
			log.Warn(fmt.Sprintf("Unable to get catalog for service [%s]", cf.ConfigOrigin.ServiceName))
		}
	}
	if len(svcs) == 0 {
		if cf.ConfigOrigin.IsFatal == true {
			panic(fmt.Sprintf("No service [%s] available", cf.ConfigOrigin.ServiceName))
		} else {
			log.Warn(fmt.Sprintf("No service [%s] available", cf.ConfigOrigin.ServiceName))
		}
	}

	// Fetch Config
	for _, svc := range svcs {
		r := *http.DefaultClient
		log.Debug(fmt.Sprintf("Fetching http://%s:%d/%s", svc.Address, svc.ServicePort, cf.ConfigOrigin.Path))
		resp, err := r.Get(fmt.Sprintf("http://%s:%d/%s", svc.Address, svc.ServicePort, cf.ConfigOrigin.Path))
		if err != nil {
			log.Warn(err)
			continue
		}
		defer resp.Body.Close()

		// Unmarshall Yaml
		b := make([]byte, resp.ContentLength)
		resp.Body.Read(b)
		t := RemoteConfig{}
		err = yaml.Unmarshal(b, &t)
		if err != nil {
			log.Warn(err)
			continue
		}
		return t
	}
	if cf.ConfigOrigin.IsFatal == true {
		panic("No config available")
	}
	log.Warn("No config available")
	return RemoteConfig{}

}

func NewConfigFetcher(cc *ConsulConf, co *ConfigOrigin) *ConfigFetcher {
	return &ConfigFetcher{
		ConsulApi:    nil,
		ConsulConf:   cc,
		ConfigOrigin: co,
	}
}
