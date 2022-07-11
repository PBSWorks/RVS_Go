package toc

import (
	"altair/rvs/datamodel"
	l "altair/rvs/globlog"
	"encoding/json"
	"io/ioutil"
	"os"
)

var plot *datamodel.Plots

func readTOCAndWriteFilterTOC(plottocfile string) string {

	jsonFile, err := os.Open(plottocfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		l.Log().Error(err)
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
	return string(filterTOCstring)
}
