package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type WebServer struct {
	Address             string
	Port                int
	RegisterToConsul    bool
	ConsulApi           *api.Client
	ConsulAgent         *api.Agent
	ServiceRegistration *api.AgentServiceRegistration
	Listener            net.Listener
	Config              *WebServerConfiguration
	RootFolder          string
	err                 error
}

func NewWebServer(s *WebServerConfiguration) *WebServer {
	p := new(WebServer)
	p.Config = s
	p.Address = s.EntryPoint.Address
	p.Port = s.EntryPoint.Port
	p.RegisterToConsul = s.Consul.Register
	p.ConsulApi = nil
	p.ConsulAgent = nil
	p.Listener = nil
	return p
}

func (server *WebServer) AddExitHook() {
	/* Capture Signal */
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Printf("captured %v, stopping profiler and exiting..", sig)
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
	server.Port = server.Listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Binded port: %d\n", server.Port)
	if server.RegisterToConsul == true {
		server.RegisterConsul()
	}
	server.AddExitHook()
	return true
}

func (server *WebServer) RegisterConsul() bool {
	server.ConsulApi, server.err = api.NewClient(api.DefaultConfig())
	if server.err != nil {
		panic("unable to connect to consul")
	}
	server.ConsulAgent = server.ConsulApi.Agent()

	server.ServiceRegistration = &api.AgentServiceRegistration{
		ID:                "support-4242",
		Name:              server.Config.ServiceName,
		Tags:              server.Config.Consul.Tags,
		Port:              server.Port,
		Address:           server.Address,
		EnableTagOverride: false,
		Checks:            []*api.AgentServiceCheck{}}

	err := server.ConsulAgent.ServiceRegister(server.ServiceRegistration)
	if err != nil {
		panic("Unable to register to Consul")
	}
	return true
}

func (server *WebServer) Run() {
	fmt.Println(server.RootFolder)
	srv := http.Serve(server.Listener, http.FileServer(http.Dir(server.RootFolder)))
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
