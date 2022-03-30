package main

import (
	"io/ioutil"
	"testing"
)

func readHtmlFromFile(fileName string) (string, error) {

	bs, err := ioutil.ReadFile(fileName)

	if err != nil {
			return "", err
	}

	return string(bs), nil
}

func TestOccupancyExtraction(t *testing.T) {
	index, err := readHtmlFromFile("../test/resources/techlib.index.html")

	if err != nil {
		t.Fatal("Can't read source html")
	}

	occupancy := extractCurrentNumberOfPersons(index)
	if occupancy != 170 {
		t.Errorf("Invalid occupancy %d != 170", occupancy)
	}
	
}
