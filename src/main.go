package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

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

func main() {
	log.Println("Starting app")

	techlibIndex, err := getTechlibIndex()
	if err != nil {
		log.Panic("Err")
	}

	currentOccupancy := extractCurrentNumberOfPersons(techlibIndex) 


	fmt.Printf("Current -> %d\n", currentOccupancy)
}
