package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
	"golang.org/x/net/html"
)

const currentOccupancyText = "Current\u00A0number of\u00A0persons in the library"
const techlibEnIndex = "https://www.techlib.cz/en/"

func extractCurrentNumberOfPersons(techlibIndex string) int {
	log.Println("Extracting current occupancy from techlib index page")
	tkn := html.NewTokenizer(strings.NewReader(techlibIndex))
	
	var occupancyTextReached bool
	var isSpan bool
	
	for {
		tt := tkn.Next()

		switch tt {
		case html.ErrorToken:
			log.Println("Hit a error token, skipping further processing")
			return -1

		case html.StartTagToken:
			t := tkn.Token()
			isSpan = (t.Data) == "span"

		case html.TextToken:
			t := tkn.Token()

			if strings.HasPrefix(t.Data, currentOccupancyText) {
				occupancyTextReached = true
			} else if occupancyTextReached && isSpan {				
				trimmed := strings.TrimSpace(t.Data)
				occupancy, err := strconv.Atoi(trimmed)

				if err != nil {
					log.Printf("Can't convert %s to int. Skipping further processing\n", trimmed)
					return -2
				} else {
					log.Printf("Current occupancy successfully extracted => %d\n", occupancy)
					return occupancy
				}
			}
			isSpan = false
		}
	}
}

func getTechlibIndex() (string, error) {
	resp, err := http.Get(techlibEnIndex)
	if err != nil {
		log.Printf("GET %s failed\n", techlibEnIndex) 
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Reading of requested data failed!")
		return "", err
	}

	return string(body), nil
}

func writeCurrentoccupancyToInflux(currOccupancy int, influxHost string, authToken string) bool {
	log.Printf("Connecting to influx host %s\n", influxHost)
	client := influxdb2.NewClient(influxHost, authToken)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking("io.jacfal", "utils")
	p := influxdb2.NewPointWithMeasurement("techlib-occupancy").
		AddTag("unit", "occ").
		AddField("value", currOccupancy).
		SetTime(time.Now())
	
	err := writeAPI.WritePoint(context.Background(), p)
	if (err != nil) {
		log.Printf("InfluxDB writting error: %s\n", err)
		return false
	}

	return true 
}

func main() {
	log.Println("Starting app")

	influxdbHost := os.Getenv("INFLUXDB_HOST")
	influxdbAuthToken := os.Getenv("INFLUXDB_AUTH")

	log.Printf("Using influxdb host %s\n", influxdbHost)

	if influxdbHost == "" || influxdbAuthToken == "" {
		log.Panic("INFLUXDB_HOST or INFLUXDB_AUTH env var missing")
	}

	for {
		time.Sleep(1 * time.Minute)
		log.Println("Starting new cycle...")

		techlibIndex, err := getTechlibIndex()
		if err != nil {
			log.Printf("Can't get techlib eng index -> %s", err)
			continue
		}

		currentOccupancy := extractCurrentNumberOfPersons(techlibIndex) 
		writeSuccess := writeCurrentoccupancyToInflux(currentOccupancy, influxdbHost, influxdbAuthToken)
		if writeSuccess {
			log.Println("Writting success")
		} else {
			log.Println("Writting failed!")
		}
	}
}
