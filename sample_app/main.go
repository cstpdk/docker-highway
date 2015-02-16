package main

import (
	"net/http"
	"fmt"
	"os"
	"io/ioutil"
)

func main() {

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {

		resp, _ := http.Get("http://curlmyip.com")
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		fmt.Fprintf(w, "Hallo welt, aus %s",string(body))
	})

	http.ListenAndServe(fmt.Sprintf(":%s",os.Getenv("PORT")), nil)
}
