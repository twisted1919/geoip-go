package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/oschwald/geoip2-golang"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// main configuration strct
type configuration struct {
	IP           string `json:"server.ip"`
	Port         int    `json:"server.port"`
	Password     string `json:"server.password"`
	DatabaseFile string `json:"database.file"`
}

// create a new configuration with default values
func newConfiguration() *configuration {
	return &configuration{
		IP:           "127.0.0.1",
		Port:         8000,
		Password:     "",
		DatabaseFile: "",
	}
}

func (c *configuration) loadFromJSONFile(configFile string) {
	currentPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	configFilePath := currentPath + string(os.PathSeparator) + configFile

	_, err = os.Stat(configFilePath)
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Configuration file read error: %s", err)
	}

	err = json.Unmarshal(b, c)
	if err != nil {
		log.Fatalf("Configuration file marshal error: %s", err)
	}
}

type httpJSONResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type geoDataResponse struct {
	Continent   string  `json:"continent"`
	CountryName string  `json:"country_name"`
	CountryCode string  `json:"country_code"`
	StateName   string  `json:"state_name"`
	CityName    string  `json:"city_name"`
	PostalCode  string  `json:"postal_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	TimeZone    string  `json:"timezone"`
}

var (
	config *configuration
	db     *geoip2.Reader
)

func setupHTTP(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		fn(w, r, ps)
	}
}

func sendHTTPJSONResponse(w http.ResponseWriter, status, message string, data interface{}) {
	js, err := json.Marshal(&httpJSONResponse{status, message, data})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(js))
	return
}

func httpHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if len(config.Password) > 0 && r.Header.Get("Authorization") != config.Password {
		sendHTTPJSONResponse(w, "error", "Invalid password", nil)
		return
	}
	tStart := time.Now()

	ipAddr := ps.ByName("ip")
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		ips, err := net.LookupIP(ipAddr)
		if err != nil {
			sendHTTPJSONResponse(w, "error", "Invalid ip address", nil)
			return
		}
		ip = ips[0]
	}

	if ip == nil {
		sendHTTPJSONResponse(w, "error", "Invalid ip address", nil)
		return
	}

	record, err := db.City(ip)
	if err != nil {
		sendHTTPJSONResponse(w, "error", "Cannot process request", nil)
		return
	}

	stateName := ""
	if len(record.Subdivisions) > 0 {
		stateName = record.Subdivisions[0].Names["en"]
	}
	res := &geoDataResponse{
		Continent:   record.Continent.Names["en"],
		CountryName: record.Country.Names["en"],
		CountryCode: record.Country.IsoCode,
		StateName:   stateName,
		CityName:    record.City.Names["en"],
		PostalCode:  record.Postal.Code,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		TimeZone:    record.Location.TimeZone,
	}

	tElapsed := time.Since(tStart)
	sendHTTPJSONResponse(w, "success", fmt.Sprintf("OK [took %s]", tElapsed), res)
}

func aliveHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if len(config.Password) > 0 && r.Header.Get("Authorization") != config.Password {
		fmt.Fprint(w, "Invalid password")
		return
	}
	fmt.Fprint(w, "pong")
	return
}

func main() {

	defaultConfig := newConfiguration()
	defaultConfig.loadFromJSONFile("config.json")

	ip := flag.String("server.ip", defaultConfig.IP, "server ip address, empty to bind all interfaces")
	port := flag.Int("server.port", defaultConfig.Port, "server port")
	password := flag.String("server.password", defaultConfig.Password, "the password to allow access to the server via http requests")
	dbFile := flag.String("database.file", defaultConfig.DatabaseFile, "the database file that contains GeoIP information")

	flag.Parse()

	config = &configuration{
		IP:           *ip,
		Port:         *port,
		Password:     *password,
		DatabaseFile: *dbFile,
	}

	// no need anymore
	defaultConfig = nil

	// database file
	var err error
	db, err = geoip2.Open(config.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	address := fmt.Sprintf("%s:%d", config.IP, config.Port)
	router := httprouter.New()
	router.GET("/ping", aliveHandler)
	router.GET("/check/:ip", setupHTTP(httpHandler))
	log.Fatal(http.ListenAndServe(address, router))
}
