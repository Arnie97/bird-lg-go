package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"

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

// Access handler, check to see if client IP in allowed IPs, continue if it is, send to invalidHandler if not
func accessHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpW http.ResponseWriter, httpR *http.Request) {

		// setting.allowedIPs will always have at least one element because of how it's defined
		if setting.allowedIPs[0] == "" {
			next.ServeHTTP(httpW, httpR)
			return
		}

		IPPort := httpR.RemoteAddr

		// Remove port from IP and remove brackets that are around IPv6 addresses
		requestIp := IPPort[0:strings.LastIndex(IPPort, ":")]
		requestIp = strings.Replace(requestIp, "[", "", -1)
		requestIp = strings.Replace(requestIp, "]", "", -1)

		for _, allowedIP := range setting.allowedIPs {
			if requestIp == allowedIP {
				next.ServeHTTP(httpW, httpR)
				return
			}
		}

		invalidHandler(httpW, httpR)
		return
	})
}

type settingType struct {
	birdSocket  string
	listen      string
	allowedIPs  []string
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
		":8000",
		[]string{""},
		"",
		"templates",
	}

	if birdSocketEnv := os.Getenv("BIRD_SOCKET"); birdSocketEnv != "" {
		settingDefault.birdSocket = birdSocketEnv
	}
	if listenEnv := os.Getenv("BIRDLG_LISTEN"); listenEnv != "" {
		settingDefault.listen = listenEnv
	}
	if AllowedIPsEnv := os.Getenv("ALLOWED_IPS"); AllowedIPsEnv != "" {
		settingDefault.allowedIPs = strings.Split(AllowedIPsEnv, ",")
	}
	if peeringEnv := os.Getenv("BIRDLG_PEERING"); peeringEnv != "" {
		settingDefault.peeringConf = peeringEnv
	}
	if templatesEnv := os.Getenv("BIRDLG_TEMPLATES"); templatesEnv != "" {
		settingDefault.templates = templatesEnv
	}

	// Allow parameters to override environment variables
	birdParam := flag.String("bird", settingDefault.birdSocket, "socket file for bird, set either in parameter or environment variable BIRD_SOCKET")
	listenParam := flag.String("listen", settingDefault.listen, "listen address, set either in parameter or environment variable BIRDLG_LISTEN")
	AllowedIPsParam := flag.String("allowed", strings.Join(settingDefault.allowedIPs, ","), "IPs allowed to access this proxy, separated by commas. Don't set to allow all IPs.")
	peeringParam := flag.String("peering", settingDefault.peeringConf, "peering config file, set either in parameter or environment variable BIRDLG_PEERING")
	templatesParam := flag.String("templates", settingDefault.templates, "peering config file, set either in parameter or environment variable BIRDLG_TEMPLATES")
	flag.Parse()

	setting.birdSocket = *birdParam
	setting.listen = *listenParam
	setting.allowedIPs = strings.Split(*AllowedIPsParam, ",")
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
	http.HandleFunc("/bird", birdHandler)
	http.HandleFunc("/bird6", birdHandler)
	http.HandleFunc("/traceroute", tracerouteHandler)
	http.HandleFunc("/traceroute6", tracerouteHandler)
	http.HandleFunc("/peering", peeringWrapper)
	http.ListenAndServe(*listenParam, handlers.LoggingHandler(os.Stdout, accessHandler(http.DefaultServeMux)))
}
