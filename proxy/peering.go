package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type PointOfPresence struct {
	ASNumberFull        uint32 `json:"asn"`
	ASNumberSuffix      string `json:"-"`
	WireGuardEndpoint   string `json:"wg"`
	WireGuardPrivateKey string `json:"priv,omitempty"`
	WireGuardPublicKey  string `json:"publ,omitempty"`
	TunneledIPv4        net.IP `json:"ipv4,omitempty"`
	TunneledIPv6        net.IP `json:"ipv6,omitempty"`
	LinkLocalIPv6       net.IP `json:"link,omitempty"`
	Name                string `json:"name,omitempty"`
	AdditionalNotes     string `json:"note,omitempty"`
}

type Communities struct {
	Latency, Bandwidth uint32
}

type Peering struct {
	Alice, Bob PointOfPresence
	Communities
}

type TemplateFile struct {
	Path              string
	FileName, Content *template.Template
}

// bob should not tell alice his private key, and vice versa
func (p *Peering) MaskPrivateKeys() *Peering {
	p.Alice.WireGuardPrivateKey = "(private key)"
	p.Bob.WireGuardPrivateKey = "(private key)"
	return p
}

func loadTemplates() {
	filepath.Walk(setting.templates, func(path string, info os.FileInfo, err error) error {
		var tmpl TemplateFile
		if err != nil {
			return err
		} else if !strings.Contains(info.Name(), "{{") {
			return nil
		} else if tmpl.FileName, err = template.New(info.Name()).Parse(info.Name()); err != nil {
			return err
		} else if bytes, err := ioutil.ReadFile(path); err != nil {
			return err
		} else if tmpl.Content, err = template.New(info.Name()).Parse(string(bytes)); err != nil {
			return err
		} else {
			tmpl.Path = path
			templates = append(templates, tmpl)
			return nil
		}
	})
}

func peeringWrapper(httpW http.ResponseWriter, httpR *http.Request) {
	if peeringConf == nil {
		invalidHandler(httpW, httpR)
		return
	}

	switch httpR.Method {
	case "GET":
		resp := *peeringConf
		json.NewEncoder(httpW).Encode(resp.MaskPrivateKeys())
	case "POST":
		var (
			err  error
			resp struct {
				Error string
				Files map[string]string
			}
		)
		resp.Files, err = peeringHandler(httpR.Body)
		if err != nil {
			resp.Error = err.Error()
		} else {
			birdReconfigure(setting.birdSocket)
		}
		json.NewEncoder(httpW).Encode(&resp)
	}
}

func peeringHandler(body io.ReadCloser) (map[string]string, error) {
	var (
		req      *Peering
		resp     = make(map[string]string)
		fileName = bytes.NewBuffer(nil)
		content  = bytes.NewBuffer(nil)
	)
	defer body.Close()
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return nil, err
	}

	// do not trust bob, use local config for alice
	localConf := &Peering{
		Alice:       peeringConf.Alice,
		Bob:         req.Bob,
		Communities: req.Communities,
	}

	// this is the only field in alice that bob might change
	if req.Alice.LinkLocalIPv6.IsLinkLocalUnicast() {
		localConf.Alice.LinkLocalIPv6 = req.Alice.LinkLocalIPv6
	}

	// swap alice and bob to generate an example config for the peer AS
	peerConf := (&Peering{
		Alice:       localConf.Bob,
		Bob:         localConf.Alice,
		Communities: localConf.Communities,
	}).MaskPrivateKeys()

	if err := getASNumberSuffix(localConf); err != nil {
		return nil, err
	}

	for _, tmpl := range templates {

		// local config
		if err := tmpl.FileName.Execute(fileName, localConf); err != nil {
			return nil, err
		} else if file, err := os.OpenFile(filepath.Join(filepath.Dir(tmpl.Path), fileName.String()), os.O_WRONLY|os.O_CREATE, 0644); err != nil {
			return nil, err
		} else if err = tmpl.Content.Execute(file, localConf); err != nil {
			file.Close()
			os.Remove(file.Name())
			return nil, err
		}
		fileName.Reset()

		// peer config
		if err := tmpl.FileName.Execute(fileName, peerConf); err != nil {
			return nil, err
		} else if err = tmpl.Content.Execute(content, peerConf); err != nil {
			return nil, err
		} else {
			resp[fileName.String()] = content.String()
		}
		fileName.Reset()
		content.Reset()
	}
	return resp, nil
}

// append an "automated peering" entry to the end of bird output
func peeringForm(query string, httpW http.ResponseWriter) {
	if peeringConf == nil || query != "show protocols" {
		return
	}
	httpW.Write([]byte(fmt.Sprintf(
		"new_peer BGP automated open %s Peer with me in a minute!\n",
		time.Now().Format("2006-01-02"),
	)))
}

var isAlphaNumeric = regexp.MustCompile(`^\w+$`).MatchString

func getASNumberSuffix(conf *Peering) error {
	for _, role := range []*PointOfPresence{&conf.Alice, &conf.Bob} {
		fullASN := strconv.FormatInt(int64(role.ASNumberFull), 10)
		role.ASNumberSuffix = ("0000" + fullASN)[len(fullASN):]
	}

	var (
		index = strings.IndexRune(conf.Alice.WireGuardEndpoint, ':')
		port  string
	)
	if index >= 0 {
		port = conf.Alice.WireGuardEndpoint[index+1:]
	} else {
		port = "2" + conf.Bob.ASNumberSuffix
	}

	if listener, err := net.ListenPacket("udp", ":"+port); err != nil {
		return err
	} else if err = listener.Close(); err != nil {
		return err
	}
	conf.Alice.WireGuardEndpoint = port

	if !isAlphaNumeric(conf.Bob.Name) {
		return fmt.Errorf("identifier '%s' should be modified to be alphanumeric", conf.Bob.Name)
	}
	return nil
}
