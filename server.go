package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
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
	Proxies             map[string]*httputil.ReverseProxy
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
	p.Proxies = make(map[string]*httputil.ReverseProxy)
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

func (server *WebServer) GetExposedIpAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic("Unable to get IPs")
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	hostname, err := os.Hostname()
	if err != nil {
		panic("Unable to get Hostname")
	}
	return hostname
}

func (server *WebServer) Init() bool {
	server.Listener, server.err = net.Listen("tcp", fmt.Sprintf(":%d", server.Port))
	if server.err != nil {
		panic("bind failed")
	}

	server.Address = server.GetExposedIpAddress()
	server.Port = server.Listener.Addr().(*net.TCPAddr).Port
	log.Info(fmt.Sprintf("Listening to %s:%d", server.Config.EntryPoint.Address, server.Port))

	// Init Name
	server.ServiceID = fmt.Sprintf("%s-%d-%s", server.Config.ServiceName, server.Port, server.GenerateUniqueID())

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
		srw := &StatusResponseWriter{ResponseWriter: w}
		srw.WriteHeader(200)
		fmt.Fprintf(srw, "I'm OK !")
	})

	// Dumping tags
	for _, v := range server.Config.ConsulTags {
		log.Info(fmt.Sprintf("[consul] Tag: %s", v))
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
		Tags:              server.Config.ConsulTags,
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
		log.Warning(err)
		panic("Unable to register to Consul")
	}
	log.Info("Registered to consul")
	return true
}

func (server *WebServer) Run() {
	// Prepare Proxified stuff
	for k, pr := range server.Config.RouteMappings {
		d := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				newUrl := req.URL.Path
				if pr.StripPath {
					newUrl = req.URL.Path[len(pr.Path):len(req.URL.Path)]
				}
				if pr.PrefixPath != "" {
					newUrl = pr.PrefixPath + newUrl
				}
				log.Info(fmt.Sprintf("[proxy][%s] %s to %s://%s/%s", k, req.URL.Path, pr.Scheme, pr.Host, newUrl))
				req.URL.Path = newUrl
				req.URL.Scheme = pr.Scheme
				req.URL.Host = pr.Host
			},
		}
		server.Proxies[pr.Path] = d
		log.Info(fmt.Sprintf("[proxy][%s] %s will be forworded to %s://%s/%s", k, pr.Path, pr.Scheme, pr.Host, pr.PrefixPath))
	}

	// Handle Static stuff
	log.Info(fmt.Sprintf("Serving %s", server.Config.RootFolder))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		srw := &StatusResponseWriter{ResponseWriter: w}
		defer func() {
			server.LogResponse(srw, r)
		}()

		// Check for Proxies
		for k, v := range server.Proxies {
			if strings.HasPrefix(r.URL.Path, k) {
				v.ServeHTTP(srw, r)
				return
			}
		}

		// Failback on Static Files
		if server.Config.AngularMode {
			rx, err := regexp.MatchString("^/([^/]+)\\.((ttf|eot|svg|js|woff2|map|ico)(\\?.*)?)", r.URL.Path)
			if !strings.HasPrefix(r.URL.Path, "/app/") && !strings.HasPrefix(r.URL.Path, "/assets/") && err == nil && rx == false {
				r.URL.Path = "/"
			}
		}
		server.FileServer.ServeHTTP(srw, r)
	})

	// Serve
	srv := server.Server.Serve(server.Listener)
	panic(srv)
}

func (server *WebServer) LogResponse(srw *StatusResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"status":      srw.status,
		"path":        r.RequestURI,
		"method":      r.Method,
		"proto":       r.Proto,
		"remote-addr": r.RemoteAddr,
	}).Info(fmt.Sprintf("[%d] %s", srw.status, r.RequestURI))
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
