package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/go-yaml/yaml"
	"github.com/hashicorp/consul/api"
	"net/http"
)

type TagsFetcher struct {
	ConsulApi  *api.Client
	ConsulConf *ConsulConf
	TagsOrigin *TagsOrigin
	Catalog    *api.Catalog
}

func (tf *TagsFetcher) Tags() []string {
	var err error = nil

	log.Info(fmt.Sprintf("Loading tags from %s on %s:%d",
		tf.TagsOrigin.ServiceName,
		tf.ConsulConf.Host,
		tf.ConsulConf.Port))

	// Connect to consul
	tf.ConsulApi, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		panic("Unable to connect to consul")
	}

	// Fetch Service Catalog
	tf.Catalog = tf.ConsulApi.Catalog()
	svcs, _, err := tf.Catalog.Service(tf.TagsOrigin.ServiceName, "", nil)
	if err != nil {
		if tf.TagsOrigin.IsFatal == true {
			panic("Unable to get catalog")
		} else {
			log.Warn(fmt.Sprintf("Unable to get catalog for service [%s]", tf.TagsOrigin.ServiceName))
		}
	}
	if len(svcs) == 0 {
		if tf.TagsOrigin.IsFatal == true {
			panic(fmt.Sprintf("No service [%s] available", tf.TagsOrigin.ServiceName))
		} else {
			log.Warn(fmt.Sprintf("No service [%s] available", tf.TagsOrigin.ServiceName))
		}
	}

	// Fetch Config
	for _, svc := range svcs {
		r := *http.DefaultClient
		log.Debug(fmt.Sprintf("Fetching http://%s:%d/%s", svc.Address, svc.ServicePort, tf.TagsOrigin.Path))
		resp, err := r.Get(fmt.Sprintf("http://%s:%d/%s", svc.Address, svc.ServicePort, tf.TagsOrigin.Path))
		if err != nil {
			log.Warn(err)
			continue
		}
		defer resp.Body.Close()

		// Unmarshall Yaml
		b := make([]byte, resp.ContentLength)
		resp.Body.Read(b)
		t := FetchedTags{}
		err = yaml.Unmarshal(b, &t)
		if err != nil {
			log.Warn(err)
			continue
		}
		return t.Tags
	}
	if tf.TagsOrigin.IsFatal == true {
		panic("No tags available")
	}
	log.Warn("No tags available")
	return []string{}

}

func NewTagsFetcher(cc *ConsulConf, to *TagsOrigin) *TagsFetcher {
	return &TagsFetcher{
		ConsulApi:  nil,
		ConsulConf: cc,
		TagsOrigin: to,
	}
}
