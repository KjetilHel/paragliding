package main

import (
	"encoding/json"
	"github.com/marni/goigc"
	"google.golang.org/appengine"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var startTime time.Time // Stopwatch keeping track of uptime
var results []int       // An array containing all the ID's of the igc's on the site
var igcs []string       // An array containing the url for all the igc's on the site
var idCount int         // A counter for incrementing igc IDs so every igc gets an unique ID

// Response is struct for the response after a new igc is POSTed to the site
type Response struct {
	ID int `json:"id"`
}

// Post is the struct for decoding an incoming POST-request
type Post struct {
	URL string `json:"url"`
}

// IgcInfo is the struct fo all the wanted data from a igc file
type IgcInfo struct {
	HDate       string  `json:"h_date"`
	Pilot       string  `json:"pilot"`
	Glider      string  `json:"glider"`
	GiderID     string  `json:"glider_id"`
	TrackLength float64 `json:"track_length"`
	TrackSrcUrl string 	`json:"track_src_url"`
}

// APIInfo is the struct for the api-information
type APIInfo struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

// Returns the time since the service was deployed
func uptime() time.Duration {
	return time.Since(startTime)
}

// init is called when service is deployed
func init() {
	startTime = time.Now() // Starts the timer
	idCount = 0            // Initialises the ID count to 0
}

func main() {
	// The service handles three different patterns:
	// /igcinfo/api
	// /igcinfo/api/igc
	// /igcinfo/api/igc/*
	http.HandleFunc("/paragliding/api", infoHandler)
	http.HandleFunc("/paragliding/", igcHandler)

	appengine.Main() // Required for the service to work on GoogleCloud

	// Connects the service to a port and listens to that port
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal(err)
	}

}

// Writes info about the service to the screen
func infoHandler(w http.ResponseWriter, r *http.Request) {
	api := APIInfo{uptime().String(), "Service for IGC tracks.", "v1"}
	err := json.NewEncoder(w).Encode(api)
	if err != nil {
		panic(err)
	}
}

func igcHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// Handles when a igc is posted to the service
	if r.Method == "POST" {
		if len(url) == 4 {
			// Reads all the parameters in the POST
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			// Decodes the parameters into an Post-struct
			var p Post
			err = json.Unmarshal(body, &p)
			if err != nil {
				panic(err)
			}

			// Adds the url to the url-array and the id to the id-array
			igcs = append(igcs, p.URL)
			results = append(results, idCount)

			// Sends the id as a response to the client
			r := Response{idCount}
			err = json.NewEncoder(w).Encode(r)
			if err != nil {
				panic(err)
			}

			// Increments the counter
			idCount++
		}
	} else if r.Method == "GET" {
		if len(url) == 4 {
			// Prints out the array with ID's
			err := json.NewEncoder(w).Encode(results)
			if err != nil {
				panic(err)
			}
		} else if len(url) == 5 {
			switch url[3] {
			case "track":
			case "ticker":
			}
		}
	}
}

// Calculates the total distance of a igc-track
func distOfTrack(p []igc.Point) float64 {
	totaldist := 0.0
	for i := 1; i < len(p); i++ {
		totaldist += (*igc.Point).Distance(&p[i], p[i-1])
	}
	return totaldist
}
