package graph

import (
	"altair/rvs/datamodel"
	"encoding/json"
	"io/ioutil"
)

func writeToPLTFile(temporaryPLTFile string, lstPlotRequestResponseModel []datamodel.PlotRequestResponseModel) {
	writeToInstanceFile(temporaryPLTFile, lstPlotRequestResponseModel)
}

func writeToInstanceFile(temporaryPLTFile string,
	lstPlotRequestResponseModel []datamodel.PlotRequestResponseModel) {

	var Plotinstance datamodel.Plotinstance
	if len(lstPlotRequestResponseModel) > 0 {

		for i := 0; i < len(lstPlotRequestResponseModel); i++ {
			Plotinstance.Instances.PLT =
				append(Plotinstance.Instances.PLT, datamodel.PLT{
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
