package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {

	http.HandleFunc("/", echoHandler)

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <port1> <port2> ... <portN>\n", os.Args[0])
		return
	}

	for _, r := range os.Args[1:] {
		_, err := strconv.Atoi(r)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}
		go func(port string) {
			fmt.Printf("Listening on port %s\n", port)
			log.Fatal(http.ListenAndServe(":"+port, nil))
		}(r)
	}
	select {}

}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request received: %s\n", r.URL.String())
	res := make(map[string]interface{})

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	if len(body) > 0 {
		res["body"] = string(body)
	}
	res["method"] = r.Method
	res["url"] = r.URL.String()
	res["header"] = r.Header
	res["host"] = r.Host
	res["remoteAddr"] = r.RemoteAddr
	res["requestURI"] = r.RequestURI
	res["proto"] = r.Proto
	res["contentLength"] = r.ContentLength
	res["transferEncoding"] = r.TransferEncoding
	res["close"] = r.Close
	res["form"] = r.Form
	res["postForm"] = r.PostForm
	res["multipartForm"] = r.MultipartForm
	res["trailer"] = r.Trailer
	res["tls"] = r.TLS

	for k, v := range res {
		if v == nil || v == "" {
			delete(res, k)
		}
	}

	str, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	fmt.Fprintf(w, "%s", str)
}
