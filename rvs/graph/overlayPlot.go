package graph

import (
	"altair/rvs/common"
	"altair/rvs/exception"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type plotOverlayModel struct {
	TemporaryPltFilePath         string                       `json:"temporaryPltFilePath"`
	OverlaidPlotFileRemovedModel overlaidPlotFileRemovedModel `json:"overlaidPlotFileRemovedModel"`
	OverlaidPlotFile             resultFileInformationModel   `json:"overlaidPlotFile"`
	OriginalResultFile           resultFileInformationModel   `json:"originalResultFile"`
	PlotSelectionModelList       plotSelectionModelList       `json:"plotSelectionModelList"`
}

type overlaidPlotFileRemovedModel struct {
	OverlaidPlotFileLabelIndex      int    `json:"overlaidPlotFileLabelIndex"`
	OverlaidPlotPltBlocksStartIndex int    `json:"overlaidPlotPltBlocksStartIndex"`
	OverlaidPlotPltBlocksEndIndex   int    `json:"overlaidPlotPltBlocksEndIndex"`
	Filepath                        string `json:"filepath"`
}
type plotSelectionModelList struct {
	PlotSelectionModelList []PlotSelectionModel       `json:"plotSelectionModelList"`
	FileInformationModel   resultFileInformationModel `json:"fileInformationModel"`
}

type PlotSelectionModel struct {
	FileName        string `json:"fileName"`
	PlotTitle       string `json:"plotTitle"`
	UniquePltNumber int    `json:"uniquePltNumber"`
	FilePath        string `json:"filePath"`
}

var OverlayOriginalPlotinstance *plotinstance
var OverlayPlotinstance *plotinstance

func OverlayPlt(sRequestData []byte, username string, password string, sToken string) (string, error) {
	var PlotOverlayModel plotOverlayModel
	var newlyAddedPltBlocksCount int
	json.Unmarshal(sRequestData, &PlotOverlayModel)
	OverlayOriginalPlotinstance = new(plotinstance)
	OverlayPlotinstance = new(plotinstance)

	var overlaidFilePath = PlotOverlayModel.OverlaidPlotFile.FilePath
	var overlaidPlotModelList []plotRequestResponseModel
	var originalPlotModelList []plotRequestResponseModel

	jsonFile, err := os.Open(PlotOverlayModel.TemporaryPltFilePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &OverlayOriginalPlotinstance)

	for index := 0; index < len(OverlayOriginalPlotinstance.Instances.PLT); index++ {
		originalPlotModelList = append(originalPlotModelList, plotRequestResponseModel{
			Queries:             OverlayOriginalPlotinstance.Instances.PLT[index].Queries,
			PlotMetaData:        OverlayOriginalPlotinstance.Instances.PLT[index].PlotMetaData,
			Responses:           OverlayOriginalPlotinstance.Instances.PLT[index].Responses,
			WindowPositionModel: OverlayOriginalPlotinstance.Instances.PLT[index].WindowPositionModel})

	}

	if overlaidFilePath != "" {
		var isPLTFile = true
		if strings.HasSuffix(overlaidFilePath, ".plt") ||
			strings.HasSuffix(overlaidFilePath, ".rvs") {
			var sSelectedPLTContent = common.DownloadFileWLM(PlotOverlayModel.OverlaidPlotFile.PasUrl,
				PlotOverlayModel.OverlaidPlotFile.JobState, PlotOverlayModel.OverlaidPlotFile.JobId, PlotOverlayModel.OverlaidPlotFile.FilePath,
				sToken)

			json.Unmarshal([]byte(sSelectedPLTContent), &OverlayPlotinstance)

			for _, pltlst := range OverlayPlotinstance.Instances.PLT {
				var PlotRequestResponseModel plotRequestResponseModel
				PlotRequestResponseModel.Queries = pltlst.Queries
				PlotRequestResponseModel.PlotMetaData = pltlst.PlotMetaData
				PlotRequestResponseModel.Responses = pltlst.Responses
				PlotRequestResponseModel.WindowPositionModel = pltlst.WindowPositionModel
				overlaidPlotModelList = append(overlaidPlotModelList, PlotRequestResponseModel)
			}
			newlyAddedPltBlocksCount = len(overlaidPlotModelList)
			setQueryCounter(originalPlotModelList[len(originalPlotModelList)-1],
				overlaidPlotModelList)

			originalPlotModelList = append(originalPlotModelList, overlaidPlotModelList...)

		} else {
			isPLTFile = false
		}
		if !isPLTFile {
			var cmPlotRequestResponseModel PlotRequestResModel
			cmPlotRequestResponseModel.ResultFileInformationModel = PlotOverlayModel.OverlaidPlotFile
			var overlaidQueries queries
			var lstQueryCTypes = originalPlotModelList[0].Queries.Query
			for _, quryObj := range lstQueryCTypes {
				var obj query

				obj.ResultDataSourceRef = append(obj.ResultDataSourceRef, quryObj.ResultDataSourceRef...)
				obj.OutputSource = quryObj.OutputSource
				obj.RvpPlotDataQuery = quryObj.RvpPlotDataQuery
				obj.PlotResultQuery = quryObj.PlotResultQuery
				obj.AnimationResultQuery = quryObj.AnimationResultQuery
				obj.Session = quryObj.Session
				obj.Id = quryObj.Id
				obj.VarName = quryObj.VarName
				obj.IsCachingRequired = quryObj.IsCachingRequired

				overlaidQueries.Query = append(overlaidQueries.Query, obj)
			}
			var PlotRequestResponseModel plotRequestResponseModel
			PlotRequestResponseModel.Queries = overlaidQueries
			PlotRequestResponseModel.PlotMetaData = originalPlotModelList[0].PlotMetaData
			PlotRequestResponseModel.WindowPositionModel = originalPlotModelList[0].WindowPositionModel
			cmPlotRequestResponseModel.PlotRequestResponseModel = PlotRequestResponseModel
			var indexValue = common.GetUniqueRandomIntValue()
			var plotQueries = buildPlotQueries(cmPlotRequestResponseModel, sToken, indexValue)

			if isRVPPlotQuery(plotQueries.Query[0]) {
				// responses = m_rmPortalService.getRVPPlot(plotQueries, coreConnectorModel, httpHeaders,
				// 	sToken)
			} else {

				/*
				 * * Remove simulation as part of query in
				 */
				var lstQueries = plotQueries.Query
				for _, query := range lstQueries {
					query.PlotResultQuery.DataQuery.SimulationFilter = simulationFilter{}
				}
				var responses = getNativePlot(plotQueries.Query, plotQueries.ResultDataSource[0], cmPlotRequestResponseModel.ResultFileInformationModel,
					username, password)
				if len(responses.Responses.Responselist) > 0 {
					cmPlotRequestResponseModel.PlotRequestResponseModel.Responses = responses.Responses
				} else {
					return "", &exception.RVSError{
						Errordetails: "",
						Errorcode:    "10047",
						Errortype:    "CODE_RESULT_FILE_NOT_SUPPORTED_FOR_PLOT",
					}
				}
			}

			var lstPlotRequestResponseModels []plotRequestResponseModel
			lstPlotRequestResponseModels = append(lstPlotRequestResponseModels, cmPlotRequestResponseModel.PlotRequestResponseModel)
			setQueryCounter(originalPlotModelList[len(originalPlotModelList)-1], lstPlotRequestResponseModels)
			originalPlotModelList = append(originalPlotModelList, cmPlotRequestResponseModel.PlotRequestResponseModel)

			// one new block added due to dropped result file
			newlyAddedPltBlocksCount = 1

		}

	} else {

		var startIndex = PlotOverlayModel.OverlaidPlotFileRemovedModel.OverlaidPlotPltBlocksStartIndex
		var endIndex = PlotOverlayModel.OverlaidPlotFileRemovedModel.OverlaidPlotPltBlocksEndIndex
		var tempPlotModelList []plotRequestResponseModel
		//filepath = "file="+ overlayPlotModel.getOverlaidPlotFileRemovedModel().getFilepath();
		//action =  Action.DELETE;
		for index := 0; index < len(originalPlotModelList); index++ {
			if index < startIndex || index > endIndex {
				tempPlotModelList = append(tempPlotModelList, originalPlotModelList[index])
			}
		}

		originalPlotModelList = tempPlotModelList
	}

	var PlotRequestResModeloutput = CreatePlotResponseModel(PlotOverlayModel.OriginalResultFile, originalPlotModelList,
		common.GetPlatformIndependentFilePath(filepath.Dir(PlotOverlayModel.TemporaryPltFilePath), false), newlyAddedPltBlocksCount,
		sToken, "FROM_OVERLAY")

	PlotRequestResModeloutput.PlotRequestResponseModel.PlotMetaData.UserPreferece = GetUserPlotPreferences()

	var matchingFileListData = GetWLMFileList(PlotOverlayModel.OriginalResultFile, sToken,
		PlotOverlayModel.OriginalResultFile.FilePath, PlotOverlayModel.OriginalResultFile)
	PlotRequestResModeloutput.LstMatchingResultFiles = matchingFileListData

	if xmlstring, err := json.MarshalIndent(PlotRequestResModeloutput, "", "    "); err == nil {
		return string(xmlstring), nil
	}

	return "", nil

}

func setQueryCounter(lastModelList plotRequestResponseModel, overlaidPlotModelList []plotRequestResponseModel) {
	var xCounter = getXQueryCounter(lastModelList.Queries.Query)
	var yCounter = getYQueryCounter(lastModelList.Queries.Query)
	for _, plotRequestResponseModel := range overlaidPlotModelList {
		var lstQueryCTypes = plotRequestResponseModel.Queries.Query
		var xQueryCType = lstQueryCTypes[0]
		xQueryCType.VarName = "X" + strconv.Itoa(xCounter)
		for i := 1; i < len(lstQueryCTypes); i++ {
			lstQueryCTypes[i].VarName = "Y" + strconv.Itoa(yCounter)
			yCounter++
		}
	}
}

func getXQueryCounter(lstQueryList []query) int {
	var lastQueryNumber = 1
	if len(lstQueryList) > 0 {
		var queryModel = lstQueryList[0]
		var varName = queryModel.VarName

		varName = strings.Replace(varName, "X", "", -1)
		lastQueryNumber, _ = strconv.Atoi(varName)
		lastQueryNumber++
	}
	return lastQueryNumber
}

func getYQueryCounter(lstQueryList []query) int {
	var lastQueryNumber = 1
	if len(lstQueryList) > 0 {
		var queryModel = lstQueryList[len(lstQueryList)-1]
		var varName = queryModel.VarName
		varName = strings.Replace(varName, "Y", "", -1)

		lastQueryNumber, _ = strconv.Atoi(varName)
		lastQueryNumber++
	}
	return lastQueryNumber
}
