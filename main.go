package main

import (
	"encoding/json"
	"fmt"
	"github.com/marni/goigc"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
	TrackSrcURL string  `json:"track_src_url"`
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

	db, err := Connect()

	if db != true {
		panic("GODAMN THAT DB")
	}
	// The service handles three different patterns:
	// /igcinfo/api
	// /igcinfo/api/igc
	// /igcinfo/api/igc/*
	http.HandleFunc("/paragliding/api", infoHandler)
	http.HandleFunc("/paragliding/", igcHandler)

	//appengine.Main() // Required for the service to work on GoogleCloud

	// Connects the service to a port and listens to that port
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}

// Writes info about the service to the screen
func infoHandler(w http.ResponseWriter, r *http.Request) {
	api := APIInfo{uptime().String(), "Service for paragliding tracks.", "v1"}
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

			var i IgcInfo
			// Finds the matching url and the correct data
			data, err := igc.ParseLocation(p.URL)
			if err != nil {
				panic(err)
			}
			// Puts the data into an struct and showing it to the user
			i.HDate = data.Date.String()
			i.Pilot = data.Pilot
			i.Glider = data.GliderType
			i.GiderID = data.GliderID
			i.TrackLength = distOfTrack(data.Points)
			i.TrackSrcURL = p.URL
			addTrack(i)

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
				id, err := strconv.Atoi(url[4])
				if err != nil {
					panic(err)
				}
				if id < 0 {
					fmt.Fprintln(w, "Id can not be a negative number")
				} else if id < len(results) {
					var i IgcInfo
					// Finds the matching url and the correct data
					data, err := igc.ParseLocation(igcs[id])
					if err != nil {
						panic(err)
					}
					// Puts the data into an struct and showing it to the user
					i.HDate = data.Date.String()
					i.Pilot = data.Pilot
					i.Glider = data.GliderType
					i.GiderID = data.GliderID
					i.TrackLength = distOfTrack(data.Points)
					i.TrackSrcURL = igcs[id]
					err = json.NewEncoder(w).Encode(i)
					if err != nil {
						panic(err)
					}
				} else {
					fmt.Fprintln(w, "Id was too big")
				}
			case "ticker":
			default:
				fmt.Fprintln(w, "Not a valid field")
			}

		} else if len(url) == 6 {
			switch url[3] {
			case "track":
				url := strings.Split(r.URL.Path, "/")
				id, err := strconv.Atoi(url[4])
				if err != nil {
					panic(err)
				}
				if id < 0 {
					fmt.Fprintln(w, "Id can not be a negative number")
				} else if id < len(results) {
					data, err := igc.ParseLocation(igcs[id])
					if err != nil {
						panic(err)
					}
					field := url[5] // Switch based on the field the user wants
					switch field {
					case "pilot":
						err = json.NewEncoder(w).Encode(data.Pilot)
						if err != nil {
							panic(err)
						}
					case "glider":
						err = json.NewEncoder(w).Encode(data.GliderType)
						if err != nil {
							panic(err)
						}
					case "glider_id":
						err = json.NewEncoder(w).Encode(data.GliderID)
						if err != nil {
							panic(err)
						}
					case "track_length":
						err = json.NewEncoder(w).Encode(distOfTrack(data.Points))
						if err != nil {
							panic(err)
						}
					case "H_date":
						err = json.NewEncoder(w).Encode(data.Date.String())
						if err != nil {
							panic(err)
						}
					case "track_src_url":
						err = json.NewEncoder(w).Encode(igcs[id])
						if err != nil {
							panic(err)
						}
					default:
						fmt.Fprintln(w, "Not a valid field")
					}
				} else {
					fmt.Fprintln(w, "Id was too big")
				}
			case "ticker":
			default:
				fmt.Fprintln(w, "Not a valid field")
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
