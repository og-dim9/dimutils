package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	targetHost   = "localhost" //"svc.cluster.local"
	targetScheme = "http"
	targetPort   = 8888
	sourcePort   = 8889 //8888
	basicAuth    = ""
	basicAuthDec = ""
	direction    = "egress"
	handlers     = map[string]http.HandlerFunc{
		"ingress": ingressHandler,
		"egress":  egressHandler,
	}
	proxyPool = map[string]*httputil.ReverseProxy{}
)

func main() {
	// 🤮
	readEnvironment()

	handler := handlers[direction]
	if handler == nil {
		panic("Invalid direction: " + direction)
	}
	fmt.Println("Listening on port: " + strconv.Itoa(sourcePort))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(sourcePort), handler))
}

func egressHandler(w http.ResponseWriter, r *http.Request) {

	url := &url.URL{
		Scheme: targetScheme,
		Host:   targetHost + ":" + strconv.Itoa(targetPort),
		Path:   "/" + r.Host,
	}

	fmt.Println("URL: ", url)
	if basicAuth != "" {
		r.Header.Add("Authorization", "Basic "+basicAuth)
	}

	getProxy(url).ServeHTTP(w, r)

}

func ingressHandler(w http.ResponseWriter, r *http.Request) {

	if basicAuthDec != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user+":"+pass != basicAuthDec {
			w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "No basic auth present"}`))
			return
		}
		// remove the Authorization header before forwarding the request
		r.Header.Del("Authorization")
	}

	path := strings.Split(r.URL.Path, "/")

	url := &url.URL{
		Scheme: targetScheme,
		Host:   path[1] + "." + targetHost + ":" + strconv.Itoa(targetPort),
	}
	r.URL.Path = "/" + strings.Join(path[2:], "/")
	r.Host = url.Host
	fmt.Println("URL: ", url)

	getProxy(url).ServeHTTP(w, r)
}

func getProxy(url *url.URL) *httputil.ReverseProxy {
	//FIXME: is this a premature optimization?
	//       this was done because there was a delay in the first request,
	//		 then again after a GC run
	proxy := proxyPool[url.String()]
	if proxy == nil {
		fmt.Println("Creating new proxy for: ", url.String())
		proxy = httputil.NewSingleHostReverseProxy(url)
		proxyPool[url.String()] = proxy
	}
	return proxy
}

func readEnvironment() {
	//TODO: read from cli flags too
	// 🤮
	var err error
	if os.Getenv("TARGET_HOST") != "" {
		targetHost = os.Getenv("TARGET_HOST")
	}
	if os.Getenv("TARGET_SCHEME") != "" {
		targetScheme = os.Getenv("TARGET_SCHEME")
		if targetScheme != "http" && targetScheme != "https" {
			fmt.Println("Invalid scheme: " + targetScheme)
			panic("Invalid scheme: " + targetScheme)
		}
	}
	if os.Getenv("TARGET_PORT") != "" {
		targetPort, err = strconv.Atoi(os.Getenv("TARGET_PORT"))
		if err != nil {
			fmt.Println("Invalid port number TARGET_PORT: " + os.Getenv("TARGET_PORT"))
			panic(err)
		}
	}
	if os.Getenv("SOURCE_PORT") != "" {
		sourcePort, err = strconv.Atoi(os.Getenv("SOURCE_PORT"))
		if err != nil {
			fmt.Println("Invalid port number SOURCE_PORT: " + os.Getenv("SOURCE_PORT"))
			panic(err)
		}
	}
	if os.Getenv("BASIC_AUTH") != "" {
		basicAuthDec = os.Getenv("BASIC_AUTH")
		if len(strings.Split(basicAuthDec, ":")) != 2 {
			// base64 decode the basic auth string
			decoded, err := base64.StdEncoding.DecodeString(basicAuthDec)
			if err != nil {
				fmt.Println("Invalid basic auth string")
				panic(err)
			}
			basicAuth = basicAuthDec
			basicAuthDec = string(decoded)
		} else {
			basicAuth = base64.StdEncoding.EncodeToString([]byte(basicAuthDec))
		}
		fmt.Println("basicAuth: ", basicAuth)
		fmt.Println("basicAuthDec: ", basicAuthDec)

		fmt.Println("ingres is using basic auth:")
		fmt.Println("\tAuthorization header will be stripped from proxied requests")
	}
	if os.Getenv("DIRECTION") != "" {
		direction = os.Getenv("DIRECTION")
	}
}
