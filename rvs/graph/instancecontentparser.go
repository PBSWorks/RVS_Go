package graph

import "encoding/json"

func ParsePLTBlock(sPLTFileContent string, WindowPositionModel windowPositionModel) []plotRequestResponseModel {

	var lstPlotModel []plotRequestResponseModel
	var instances Instances

	json.Unmarshal([]byte(sPLTFileContent), &instances)

	return lstPlotModel

}
