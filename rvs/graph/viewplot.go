package graph

import (
	"altair/rvs/common"
	"encoding/json"
)

var Plotinstance *plotinstance

func ViewPLT(sRequestData []byte, pasURL string, sToken string, username string) string {

	var instanceQueryModel InstanceQueryModel
	json.Unmarshal(sRequestData, &instanceQueryModel)

	var cmPLTFileModel = instanceQueryModel.ResultFileInformationModel
	cmPLTFileModel.PasUrl = pasURL

	var sPLTFileContent = common.DownloadFileWLM(pasURL, instanceQueryModel.ResultFileInformationModel.JobState,
		instanceQueryModel.ResultFileInformationModel.JobId, instanceQueryModel.ResultFileInformationModel.FilePath, sToken)

	Plotinstance = new(plotinstance)
	var lstPlotModel []plotRequestResponseModel

	json.Unmarshal([]byte(sPLTFileContent), &Plotinstance)

	for _, pltlst := range Plotinstance.Instances.PLT {
		var PlotRequestResponseModel plotRequestResponseModel
		PlotRequestResponseModel.Queries = pltlst.Queries
		PlotRequestResponseModel.PlotMetaData = pltlst.PlotMetaData
		PlotRequestResponseModel.Responses = pltlst.Responses
		PlotRequestResponseModel.WindowPositionModel = pltlst.WindowPositionModel
		lstPlotModel = append(lstPlotModel, PlotRequestResponseModel)
	}

	var PlotRequestResModel = CreatePlotResponseModel(cmPLTFileModel, lstPlotModel,
		getDataDirectoryPath(cmPLTFileModel.ServerName, username), len(lstPlotModel), sToken, "FROM_PLT")

	PlotRequestResModel.PlotRequestResponseModel.PlotMetaData.UserPreferece = GetUserPlotPreferences()

	var matchingFileListData = GetWLMFileList(instanceQueryModel.ResultFileInformationModel, sToken,
		instanceQueryModel.ResultFileInformationModel.FilePath, instanceQueryModel.ResultFileInformationModel)
	PlotRequestResModel.LstMatchingResultFiles = matchingFileListData

	if xmlstring, err := json.MarshalIndent(PlotRequestResModel, "", "    "); err == nil {

		return string(xmlstring)
	}

	return ""

}
