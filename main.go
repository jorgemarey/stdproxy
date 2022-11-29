package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	prxy "golang.org/x/net/proxy"
)

const (
	logFileTpl = "/tmp/stdproxy_%d.log"
	authEnvVar = "PROXY_CREDS"
	version    = "0.1.0"
)

func forwardStd(conn net.Conn) {
	c := make(chan int64)

	// Read from Reader and write to Writer until EOF
	copy := func(r io.ReadCloser, w io.WriteCloser) {
		defer func() {
			r.Close()
			w.Close()
		}()
		n, _ := io.Copy(w, r)
		c <- n
	}

	go copy(conn, os.Stdout)
	go copy(os.Stdin, conn)

	b := <-c
	log.Printf("[%s]: Connection has been closed by remote peer, %d bytes has been received\n", conn.RemoteAddr(), b)
	b = <-c
	log.Printf("[%s]: Local peer has been stopped, %d bytes has been sent\n", conn.RemoteAddr(), b)
}

func getProxyAuth(credsfilename string) *prxy.Auth {
	creds, _ := os.LookupEnv(authEnvVar)
	if credsfilename != "" {
		content, err := os.ReadFile(credsfilename)
		if err != nil {
			log.Fatalf("Error reading proxy credentials file '%s': %s\n", credsfilename, err)
		}
		creds = string(content)
	}

	if creds = strings.TrimSpace(creds); creds == "" {
		log.Fatalf("Proxy credentials not found")
	}
	parts := strings.Split(creds, ":")
	if len(parts) != 2 {
		log.Fatalf("Credential format is incorrect must be '<user>:<password>'")
	}

	return &prxy.Auth{
		User:     parts[0],
		Password: parts[1],
	}
}

func proxyPass(proxy string, destination string, timeout time.Duration, auth *prxy.Auth) {
	dailer, _ := prxy.SOCKS5("tcp", proxy, auth, &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	})
	conn, err := dailer.Dial("tcp", destination)
	if err != nil {
		log.Fatalf("Error opening proxy connection: %s\n", err)
	}

	log.Println("Proxy connection opened")
	forwardStd(conn)
}

func main() {
	var (
		proxyTimeout       = flag.Duration("timeout", 5*time.Second, "Proxy connection timeout")
		basicAuth          = flag.Bool("basic-auth", false, "Use basic authentication (default is to read 'PROXY_CREDS' environment variable)")
		basicAuthCredsFile = flag.String("creds-file", "", "Filepath of proxy credentials")
		logEnable          = flag.Bool("log", false, "Enable logging")
		logfile            = flag.String("log-file", fmt.Sprintf(logFileTpl, time.Now().Unix()), "Save log execution to file")
		versionCmd         = flag.Bool("version", false, "Show program version")
	)
	flag.Parse()

	if *versionCmd {
		fmt.Printf("v%s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 3 {
		fmt.Println("Program must receive 3 arguments: 'proxyHost:proxyPort destHost destPort'")
		os.Exit(1)
	}

	log.SetFlags(0)
	logWriter := io.Discard
	if *logEnable {
		file, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Error creating log file: %s", err)
		}
		logWriter = file
	}
	log.SetOutput(logWriter)

	if *basicAuthCredsFile != "" {
		*basicAuth = true
	}

	proxy := args[0]
	destination := fmt.Sprintf("%s:%s", args[1], args[2])
	log.Printf("Proxy: %s", proxy)
	log.Printf("Proxy Timeout %d:", *proxyTimeout)
	log.Printf("Destination: %s", destination)
	log.Printf("Basic authentication: %t", *basicAuth)
	if *basicAuthCredsFile != "" {
		log.Printf("BasicAuth credentials file: '%s'\n", *basicAuthCredsFile)
	}

	var auth *prxy.Auth
	if *basicAuth {
		auth = getProxyAuth(*basicAuthCredsFile)
	}

	proxyPass(proxy, destination, *proxyTimeout, auth)
}
