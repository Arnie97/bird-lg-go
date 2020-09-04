package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

// Check if a byte is character for number
func isNumeric(b byte) bool {
	return b >= byte('0') && b <= byte('9')
}

// Default handler, returns 500 Internal Server Error
func invalidHandler(httpW http.ResponseWriter, httpR *http.Request) {
	httpW.WriteHeader(http.StatusInternalServerError)
	httpW.Write([]byte("Invalid Request\n"))
}

type settingType struct {
	birdSocket  string
	bird6Socket string
	listen      string
	peeringConf string
	templates   string
}

var (
	setting     settingType
	peeringConf *Peering
	templates   []TemplateFile
)

// Wrapper of tracer
func main() {
	// Prepare default socket paths, use environment variable if possible
	var settingDefault = settingType{
		"/var/run/bird/bird.ctl",
		"/var/run/bird/bird6.ctl",
		":8000",
		"",
		"templates",
	}

	if birdSocketEnv := os.Getenv("BIRD_SOCKET"); birdSocketEnv != "" {
		settingDefault.birdSocket = birdSocketEnv
	}
	if bird6SocketEnv := os.Getenv("BIRD6_SOCKET"); bird6SocketEnv != "" {
		settingDefault.bird6Socket = bird6SocketEnv
	}
	if listenEnv := os.Getenv("BIRDLG_LISTEN"); listenEnv != "" {
		settingDefault.listen = listenEnv
	}
	if peeringEnv := os.Getenv("BIRDLG_PEERING"); peeringEnv != "" {
		settingDefault.peeringConf = peeringEnv
	}
	if templatesEnv := os.Getenv("BIRDLG_TEMPLATES"); templatesEnv != "" {
		settingDefault.templates = templatesEnv
	}

	// Allow parameters to override environment variables
	birdParam := flag.String("bird", settingDefault.birdSocket, "socket file for bird, set either in parameter or environment variable BIRD_SOCKET")
	bird6Param := flag.String("bird6", settingDefault.bird6Socket, "socket file for bird6, set either in parameter or environment variable BIRD6_SOCKET")
	listenParam := flag.String("listen", settingDefault.listen, "listen address, set either in parameter or environment variable BIRDLG_LISTEN")
	peeringParam := flag.String("peering", settingDefault.peeringConf, "peering config file, set either in parameter or environment variable BIRDLG_PEERING")
	templatesParam := flag.String("templates", settingDefault.templates, "peering config file, set either in parameter or environment variable BIRDLG_TEMPLATES")
	flag.Parse()

	setting.birdSocket = *birdParam
	setting.bird6Socket = *bird6Param
	setting.listen = *listenParam
	setting.peeringConf = *peeringParam
	setting.templates = *templatesParam

	if setting.peeringConf != "" {
		if file, err := os.Open(setting.peeringConf); err != nil {
			panic(err)
		} else if err = json.NewDecoder(file).Decode(&peeringConf); err != nil {
			panic(err)
		} else {
			file.Close()
		}
		loadTemplates()
	}

	// Start HTTP server
	http.HandleFunc("/", invalidHandler)
	http.HandleFunc("/bird", birdIPv4Wrapper)
	http.HandleFunc("/bird6", birdIPv6Wrapper)
	http.HandleFunc("/traceroute", tracerouteIPv4Wrapper)
	http.HandleFunc("/traceroute6", tracerouteIPv6Wrapper)
	http.HandleFunc("/peering", peeringWrapper)
	http.ListenAndServe(*listenParam, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
}
