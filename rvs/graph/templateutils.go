package graph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func writeToPLTFile(temporaryPLTFile string, lstPlotRequestResponseModel []plotRequestResponseModel) {
	writeToInstanceFile(temporaryPLTFile, lstPlotRequestResponseModel)
}

func writeToInstanceFile(temporaryPLTFile string,
	lstPlotRequestResponseModel []plotRequestResponseModel) {

	var Plotinstance plotinstance
	if len(lstPlotRequestResponseModel) > 0 {

		for i := 0; i < len(lstPlotRequestResponseModel); i++ {
			fmt.Println(" lstPlotRequestResponseModel[i].Queries", lstPlotRequestResponseModel[i].Queries)
			Plotinstance.Instances.PLT =
				append(Plotinstance.Instances.PLT, PLT{
					Queries:             lstPlotRequestResponseModel[i].Queries,
					Responses:           lstPlotRequestResponseModel[i].Responses,
					PlotMetaData:        lstPlotRequestResponseModel[i].PlotMetaData,
					WindowPositionModel: lstPlotRequestResponseModel[i].WindowPositionModel,
				})
		}
	}

	if xmlstring, err := json.MarshalIndent(Plotinstance, "", "    "); err == nil {
		_ = ioutil.WriteFile(temporaryPLTFile, xmlstring, 0644)
	}
}
