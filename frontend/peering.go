package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func peeringRequest(target string, r *http.Request) (ret []byte, err error) {
	if !isValidServer(target) {
		err = fmt.Errorf("unknown server: %s", target)
		return
	}

	var (
		resp *http.Response
		url  = "http://" + target + "." + setting.domain + ":" + strconv.Itoa(setting.proxyPort) + "/peering"
	)
	switch r.Method {
	case "GET":
		resp, err = http.Get(url)
	case "POST":
		req := r.PostFormValue("json")
		fmt.Println(req)
		resp, err = http.Post(url, "application/json", strings.NewReader(req))
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if ret, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	return
}
