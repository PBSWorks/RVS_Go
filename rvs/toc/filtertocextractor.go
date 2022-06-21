package toc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var plot Plots

type Plots struct {
	Plot            Plot    `json:"Plot"`
	Animation       *string `json:"Animation"`
	Model           *string `json:"Model"`
	RvpToc          *string `json:"rvpToc"`
	SupportedPPType string  `json:"SupportedPPType"`
	Custom          *string `json:"Custom"`
}

type Plot struct {
	Subcase []Subcase `json:"Subcase"`
}

type Subcase struct {
	Index       int         `json:"index"`
	Name        string      `json:"name"`
	Simulations Simulations `json:"Simulations"`
	Type        []Type      `json:"Type"`
}
type Simulations struct {
	Count      int `json:"count"`
	IndexStart int `json:"indexStart"`
}

type Type struct {
	Index             int              `json:"index"`
	Name              string           `json:"name"`
	IsRequestFiltered bool             `json:"isRequestFiltered"`
	RequestsOverview  RequestsOverview `json:"RequestsOverview"`
	Component         []Component      `json:"Component"`
}

type RequestsOverview struct {
	StartReqName string `json:"startReqName"`
	EndReqName   string `json:"endReqName"`
	NoOfRequests int    `json:"noOfRequests"`
}

type Component struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

func readTOCAndWriteFilterTOC(plottocfile string) string {

	jsonFile, err := os.Open(plottocfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	//var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &plot)

	for i := 0; i < len(plot.Plot.Subcase); i++ {
		for j := 0; j < len(plot.Plot.Subcase[i].Type); j++ {
			plot.Plot.Subcase[i].Type[j].IsRequestFiltered = true
		}
	}

	var filterTOCstring string
	if filterTOCstring, err := json.MarshalIndent(plot, "", "    "); err == nil {
		return string(filterTOCstring)
	}
	fmt.Println(string(filterTOCstring))
	return string(filterTOCstring)
}
