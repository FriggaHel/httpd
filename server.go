package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type WebServer struct {
	ServiceID           string
	Address             string
	Port                int
	RegisterToConsul    bool
	ConsulApi           *api.Client
	ConsulAgent         *api.Agent
	ServiceRegistration *api.AgentServiceRegistration
	Listener            net.Listener
	Config              *WebServerConfiguration
	Server              http.Server
	FileServer          http.Handler
	RegexAngularMode    *regexp.Regexp
	err                 error
}

func NewWebServer(s *WebServerConfiguration) *WebServer {
	p := new(WebServer)
	p.ServiceID = ""
	p.Config = s
	p.Address = s.EntryPoint.Address
	p.Port = s.EntryPoint.Port
	p.RegisterToConsul = s.Consul.Register
	p.ConsulApi = nil
	p.ConsulAgent = nil
	p.Listener = nil
	p.Server = http.Server{}
	p.FileServer = nil
	p.RegexAngularMode = nil
	return p
}

func (server *WebServer) AddExitHook() {
	/* Capture Signal */
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Warn(fmt.Sprintf("captured %v, exiting..", sig))
			server.DeInit()
			os.Exit(0)
		}
	}()
}

func (server *WebServer) Init() bool {
	server.Listener, server.err = net.Listen("tcp", fmt.Sprintf(":%d", server.Port))
	if server.err != nil {
		panic("bind failed")
	}
	hostname, err := os.Hostname()
	if err != nil {
		panic("Unable to get Hostname")
	}
	server.Address = hostname
	server.Port = server.Listener.Addr().(*net.TCPAddr).Port
	log.Info(fmt.Sprintf("Listening to %s:%d", server.Config.EntryPoint.Address, server.Port))

	// Init Name
	server.ServiceID = fmt.Sprintf("%s-%d-%s", server.Config.ServiceName, server.Port, server.GenerateUniqueID())

	// Fetch Tags
	if server.Config.TagsOrigin.Enabled == true {
		server.Config.Consul.Tags = NewTagsFetcher(server.Config.Consul, server.Config.TagsOrigin).Tags()
	}

	if server.RegisterToConsul == true {
		server.RegisterConsul()
	}
	server.AddExitHook()

	// Static file server
	server.FileServer = http.FileServer(http.Dir(server.Config.RootFolder))
	return true
}

func (server *WebServer) RegisterConsul() bool {
	var err error = nil

	// Bind HealthCheck
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I'm OK !")
	})

	for _, t := range server.Config.Consul.Tags {
		log.Info(fmt.Sprintf("Adding tag: %s", t))
	}

	// Register to Consul
	server.ConsulApi, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		panic("Unable to connect to consul")
	}
	server.ConsulAgent = server.ConsulApi.Agent()
	server.ServiceRegistration = &api.AgentServiceRegistration{
		ID:                server.ServiceID,
		Name:              server.Config.ServiceName,
		Tags:              server.Config.Consul.Tags,
		Port:              server.Port,
		Address:           server.Address,
		EnableTagOverride: false,
		Check: &api.AgentServiceCheck{
			Interval: "10s",
			Timeout:  "1s",
			HTTP:     fmt.Sprintf("http://%s:%d/health", server.Address, server.Port),
			DeregisterCriticalServiceAfter: "15m",
		}}

	err = server.ConsulAgent.ServiceRegister(server.ServiceRegistration)
	if err != nil {
		panic("Unable to register to Consul")
	}
	return true
}

func (server *WebServer) Run() {
	log.Info(fmt.Sprintf("Serving %s", server.Config.RootFolder))

	// Handle Static stuff
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if server.Config.AngularMode {
			rx, err := regexp.MatchString("^/([^/]+)\\.((ttf|eot|svg|js|woff2|map|ico)(\\?.*)?)", r.URL.Path)

			if !strings.HasPrefix(r.URL.Path, "/app/") && !strings.HasPrefix(r.URL.Path, "/assets/") && err == nil && rx == false {
				r.URL.Path = "/"
			}
		}
		srw := &StatusResponseWriter{ResponseWriter: w}
		server.FileServer.ServeHTTP(srw, r)
		log.WithFields(log.Fields{
			"status":      srw.status,
			"path":        r.RequestURI,
			"method":      r.Method,
			"proto":       r.Proto,
			"remote-addr": r.RemoteAddr,
		}).Info(fmt.Sprintf("[%d] %s", srw.status, r.RequestURI))
	})

	// Serve
	srv := server.Server.Serve(server.Listener)
	panic(srv)
}

func (server *WebServer) UnregisterFromConsul() {
	if server.ServiceRegistration != nil {
		server.ConsulAgent.ServiceDeregister(server.ServiceRegistration.ID)
	}
}

func (server *WebServer) DeInit() {
	server.UnregisterFromConsul()
	server.Listener.Close()
}

func (server *WebServer) GenerateUniqueID() string {
	rand.Seed(time.Now().UnixNano())
	letters := "abcdef0123456789"

	b := make([]byte, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
