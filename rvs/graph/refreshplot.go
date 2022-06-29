package graph

import (
	"altair/rvs/common"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var RefreshPlotinstance *plotinstance

type plotRefreshModel struct {
	ResultFileInformationModel resultFileInformationModel `json:"resultFileInformationModel"`
	TemporaryPLTFilePath       string                     `json:"temporaryPLTFilePath"`
}

func RefreshPlt(sRequestData []byte, username string, password string, sToken string) string {
	var PlotRefreshModel plotRefreshModel
	RefreshPlotinstance = new(plotinstance)
	var lstTempPlotModel []plotRequestResponseModel
	json.Unmarshal(sRequestData, &PlotRefreshModel)

	jsonFile, err := os.Open(PlotRefreshModel.TemporaryPLTFilePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &Plotinstance)
	for _, pltlst := range Plotinstance.Instances.PLT {
		var PlotRequestResponseModel plotRequestResponseModel
		PlotRequestResponseModel.Queries = pltlst.Queries
		PlotRequestResponseModel.PlotMetaData = pltlst.PlotMetaData
		PlotRequestResponseModel.Responses = pltlst.Responses
		PlotRequestResponseModel.WindowPositionModel = pltlst.WindowPositionModel
		lstTempPlotModel = append(lstTempPlotModel, PlotRequestResponseModel)
	}
	for _, plotmodel := range lstTempPlotModel {
		plotmodel.Responses.Responselist = nil
	}

	//var plotQueries = buildPlotQueries(plotRequestResModel, token, indexValue)

	if isRVPPlotQuery(lstTempPlotModel[0].Queries.Query[0]) {

	} else {
		var lstQueries = lstTempPlotModel[0].Queries.Query
		for _, query := range lstQueries {
			query.PlotResultQuery.DataQuery.SimulationFilter = simulationFilter{}
		}

		var responses = getNativePlot(lstTempPlotModel[0].Queries.Query, lstTempPlotModel[0].Queries.ResultDataSource[0],
			PlotRefreshModel.ResultFileInformationModel, username, password)
		var lstPlotModel []plotRequestResponseModel
		var cmPlotReqResModel PlotRequestResModel

		cmPlotReqResModel.ResultFileInformationModel = PlotRefreshModel.ResultFileInformationModel
		cmPlotReqResModel.PlotRequestResponseModel = lstTempPlotModel[0]
		cmPlotReqResModel.PlotRequestResponseModel.Responses.Responselist = nil
		cmPlotReqResModel.PlotRequestResponseModel.Responses = responses.Responses

		lstPlotModel = append(lstPlotModel, cmPlotReqResModel.PlotRequestResponseModel)

		var PlotRequestResModeloutput = CreatePlotResponseModel(PlotRefreshModel.ResultFileInformationModel, lstPlotModel,
			common.GetPlatformIndependentFilePath(filepath.Dir(PlotRefreshModel.TemporaryPLTFilePath), false), 0,
			sToken, "FROM_REFRESH_PLOT")

		PlotRequestResModeloutput.PlotRequestResponseModel.PlotMetaData.UserPreferece = GetUserPlotPreferences()

		var matchingFileListData = GetWLMFileList(PlotRefreshModel.ResultFileInformationModel, sToken,
			PlotRefreshModel.ResultFileInformationModel.FilePath, PlotRefreshModel.ResultFileInformationModel)
		PlotRequestResModeloutput.LstMatchingResultFiles = matchingFileListData

		if xmlstring, err := json.MarshalIndent(PlotRequestResModeloutput, "", "    "); err == nil {
			return string(xmlstring)
		}

	}
	return ""

}
