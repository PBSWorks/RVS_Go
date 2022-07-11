package graph

import (
	"altair/rvs/datamodel"
	"encoding/json"
)

func ParsePLTBlock(sPLTFileContent string, WindowPositionModel datamodel.WindowPositionModel) []datamodel.PlotRequestResponseModel {

	var lstPlotModel []datamodel.PlotRequestResponseModel
	var instances datamodel.Instances

	json.Unmarshal([]byte(sPLTFileContent), &instances)

	return lstPlotModel

}
