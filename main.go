package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"golang.org/x/time/rate"
)

// default rate limiter
var limiter = rate.NewLimiter(1, 1)

// Remote struct to hold remote target information
type Remote struct {
	Name    string `json:"name"`    // name of the remote target
	Prefix  string `json:"prefix"`  // prefix of the remote target
	Target  string `json:"target"`  // target url
	Rewrite bool   `json:"rewrite"` // remove prefix from the request url or not
}

// Config struct to hold server configuration
type Config struct {
	Port        string   `json:"port"`          // server port
	LogPath     string   `json:"logpath"`       // log file path
	TLSVersion  string   `json:"tls_version"`   // 0.0, 1.0, 1.1, 1.2, 1.3
	CrtPath     string   `json:"crt_path"`      // certificate path
	KeyPath     string   `json:"key_path"`      // key path
	ReqLimit    int      `json:"req_limit"`     // request limit per second
	ReqBurst    int      `json:"req_burst"`     // request burst limit
	RTimeOut    int      `json:"read_timeout"`  // read timeout in seconds
	WTimeOut    int      `json:"write_timeout"` // write timeout in seconds
	IdleTimeOut int      `json:"idle_timeout"`  // idle timeout in seconds
	Targets     []Remote `json:"targets"`       // list of remote targets
}

// GetTLSVersion returns the tls version uint16 identifier based on the config
func (config *Config) GetTLSVersion() uint16 {
	switch config.TLSVersion {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return 0 // no TLS
	}
}

// limit is a middleware to limit the number of requests
func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// readConfigFile reads the configuration file and returns the Config struct
func readConfigFile() Config {
	//get config file path
	configFile := os.Getenv("YXORP_CFG_FILE")
	if configFile == "" {
		flag.StringVar(&configFile, "config", "config.json", "config file path")
		flag.Parse()
	}

	//read file
	textBytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	//parse json
	var config Config
	err = json.Unmarshal(textBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	//check that target does not have empty prefix or target
	for _, remote := range config.Targets {
		if remote.Prefix == "" || remote.Target == "" {
			log.Fatal("Remote target prefix and target url cannot be empty")
		}
	}

	//check that target does not loopback to itself
	for _, remote := range config.Targets {
		targetUrl, err := url.Parse(remote.Target)
		if err != nil {
			log.Fatal(err)
		}
		if targetUrl.Host == "localhost:"+config.Port || targetUrl.Host == "127.0.0.1:"+config.Port || targetUrl.Host == "0.0.0.0:"+config.Port {
			log.Fatal("Remote target cannot loopback to the yxorp server")
		}
	}

	return config
}

// InitLogger initializes the logger
func InitLogger(logFilePath string) func() error {
	//create log file if not exists
	_, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		file, err := os.Create(logFilePath)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}

	//open log file
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	//set multi writer
	multiWriter := io.MultiWriter(file, os.Stdout)
	log.SetOutput(multiWriter)

	return file.Close
}

func main() {
	// read config file
	config := readConfigFile()

	// initialize logger
	closeLog := InitLogger(config.LogPath)
	defer closeLog()

	// setup rate limiter
	limiter = rate.NewLimiter(rate.Limit(config.ReqLimit), config.ReqBurst)

	// create a new serve mux
	MUX := http.NewServeMux()

	// add remote targets to the mux
	for _, remote := range config.Targets {
		targetUrl, _ := url.Parse(remote.Target)
		srv := httputil.NewSingleHostReverseProxy(targetUrl)
		if remote.Rewrite {
			//if prefix does not end with /, add it
			if remote.Prefix[len(remote.Prefix)-1] != '/' {
				remote.Prefix += "/"
			}

			MUX.Handle(remote.Prefix, http.StripPrefix(remote.Prefix, srv))
		} else {
			MUX.Handle(remote.Prefix, srv)
		}
	}

	// create a new server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      limit(MUX),
		ReadTimeout:  time.Second * time.Duration(config.RTimeOut),
		WriteTimeout: time.Second * time.Duration(config.WTimeOut),
		IdleTimeout:  time.Second * time.Duration(config.IdleTimeOut),
	}

	log.Println("yxorp server is running on port " + config.Port)
	if config.GetTLSVersion() == 0 {
		// if TLS version is 0, disable TLS
		server.TLSConfig = nil
		log.Println("yxorp server is running without TLS")
		log.Fatal(server.ListenAndServe())
	} else {
		// if TLS version is set, enable TLS
		server.TLSConfig = &tls.Config{
			MinVersion: config.GetTLSVersion(),
			ServerName: "yxorp server",
		}
		log.Println("yxorp server is running with TLS version " + config.TLSVersion)

		// check if certificate and key paths are set and available
		if config.CrtPath == "" || config.KeyPath == "" {
			log.Fatal("SSL certificate and key paths must be set for TLS")
		}
		_, errCrt := os.Stat(config.CrtPath)
		if os.IsNotExist(errCrt) {
			log.Fatal("SSL certificate file does not exist")
		}

		_, errKey := os.Stat(config.KeyPath)
		if os.IsNotExist(errKey) {
			log.Fatal("SSL key file does not exist")
		}

		// start server with TLS
		log.Fatal(server.ListenAndServeTLS(config.CrtPath, config.KeyPath))
	}
}
