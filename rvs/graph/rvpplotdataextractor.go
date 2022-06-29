package graph

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type RVPProcessDataModel struct {
	RvpFileModel      datamodel.RVPFileModel `json:"rvpFileModel"`
	RvpResultFilePath string                 `json:"rvpResultFilePath"`
	ResponseFilePath  string                 `json:"responseFilePath"`
	RvpFileExtension  string                 `json:"rvpFileExtension"`
	RequestFilePath   string                 `json:"requestFilePath"`
}

type RVPPlotDataModel struct {
	PlotName        string                `json:"plotName"`
	MapColumnPoints (map[string][]string) `json:"mapColumnPoints"`
}

type TemporarySimulationQuery struct {
	StartIndex int `json:"startIndex"`
	EndIndex   int `json:"endIndex"`
	Step       int `json:"step"`
	Count      int `json:"count"`
}

func GetRVPPlotData(plotQueries queries, ResultFileInformationModel resultFileInformationModel, username string, password string) Res {
	ValidateListQueries(plotQueries, username, password)
	/*
	* Right now there will be only one query
	 */
	var rvpPlotDataQuery = plotQueries.Query[0].RvpPlotDataQuery
	/*
	* Right now we are supporting single result data source for query
	 */
	var isCachingRequired = plotQueries.Query[0].IsCachingRequired
	return readPlotData(plotQueries.ResultDataSource[0], rvpPlotDataQuery, ResultFileInformationModel, username, password, isCachingRequired)
}

func ValidateListQueries(plotQueries queries, username string, password string) error {

	var lstResultDataSource = plotQueries.ResultDataSource

	if len(lstResultDataSource) == 0 {
		var sMessage = "No Result DataSource is present"
		return &exception.RVSError{
			Errordetails: sMessage,
			Errorcode:    "5003",
			Errortype:    "TYPE_INVALID_DATASOURCE",
		}

	}

	//var lstResponseCType = new ArrayList<ResponseCType>();
	var lstQueryCType = plotQueries.Query
	for i := 0; i < len(lstQueryCType); i++ {
		/**
		 * Added for derived plot expression, This part skips validation for the queries cotain inline query and response query.
		 *
		 *
		 */
		if lstQueryCType[i].PlotResultQuery.DataQuery.InlineQuery.Title != "" {
			continue
		}

		// Check for Result DataSource Reference
		if len(lstQueryCType[i].ResultDataSourceRef) == 0 {
			var sMessage = "No Result DataSource Reference is present"

			return &exception.RVSError{
				Errordetails: sMessage,
				Errorcode:    "5003",
				Errortype:    "TYPE_INVALID_DATASOURCE",
			}
		}

		// Check for Plot result query
		if lstQueryCType[i].RvpPlotDataQuery.RvpPlotColumnInfo.ColumnName == "" {
			var sMessage = "No plot result query is present"

			return &exception.RVSError{
				Errordetails: sMessage,
				Errorcode:    "7008",
				Errortype:    "TYPE_QUERY_FAILED",
			}

		}

		var simulationQuery = lstQueryCType[i].RvpPlotDataQuery.SimulationQuery
		if simulationQuery.SimulationCountBasedQuery.Count == 0 {
			var sMessage = "No simulation query is present"
			return &exception.RVSError{
				Errordetails: sMessage,
				Errorcode:    "7008",
				Errortype:    "TYPE_QUERY_FAILED",
			}
		}

	}
	return nil

}

func readPlotData(rvpFileDataSource datamodel.ResourceDataSource,
	rvpPlotDataQuery rvpPlotDataQuery, ResultFileInformationModel resultFileInformationModel,
	username string, password string, isCachingRequired bool) Res {

	var sJobId = ResultFileInformationModel.JobId
	var sJobState = ResultFileInformationModel.JobState
	var sRVPResultFilePath string
	if sJobId == "" && sJobState == "" {
		sRVPResultFilePath, _ = common.ResolveFilePortDataSource(rvpFileDataSource, username, password)

	} else {
		sRVPResultFilePath, _ = common.ResolvePBSPortDataSource(rvpFileDataSource, username, password)

	}

	fileExtension := filepath.Ext(sRVPResultFilePath)
	var rvpFileModeldata = getRVPFileModel(fileExtension, common.RvpFilesModel)
	//TODO caching

	validateSimulationQuery(rvpPlotDataQuery.SimulationQuery)

	var sTmpRVPQueryFile = createTmpOutputFile(username, password, common.RVP_PLOT_QUERY_FILE_NAME_PART)

	if xmlstring, err := json.MarshalIndent(rvpPlotDataQuery, "", "    "); err == nil {
		f, err := os.Create(sTmpRVPQueryFile)

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		_, err2 := f.WriteString(string(xmlstring))
		if err2 != nil {
			log.Fatal(err2)
		}
	}

	var rvpProcessDataModel RVPProcessDataModel
	rvpProcessDataModel.RvpResultFilePath = sRVPResultFilePath
	rvpProcessDataModel.RvpFileModel = rvpFileModeldata
	rvpProcessDataModel.RvpFileExtension = fileExtension
	rvpProcessDataModel.RequestFilePath = sTmpRVPQueryFile

	var response = runPlotDataExtractor(rvpProcessDataModel)
	return response

}
func getRVPFileModel(fileExtension string, rvpFilesModel datamodel.SupportedRVPFilesModel) datamodel.RVPFileModel {
	var rvpFileModel datamodel.RVPFileModel
	var lstRVPFileModel = rvpFilesModel.ListRVPFileModel
	for i := 0; i < len(lstRVPFileModel); i++ {
		if strings.Contains(lstRVPFileModel[i].Pattern, fileExtension) {
			rvpFileModel = lstRVPFileModel[i]
			break
		}
	}
	return rvpFileModel
}

func validateSimulationQuery(SimulationQuery simulationQuery) error {

	if SimulationQuery.SimulationCountBasedQuery.Count == 0 && SimulationQuery.SimulationRangeBasedQuery.Step == 0 {
		return &exception.RVSError{
			Errordetails: "Simulation Query is null",
			Errorcode:    "11002",
			Errortype:    "TYPE_QUERY_FAILED",
		}
	}
	var rangeBasedQuery = SimulationQuery.SimulationRangeBasedQuery
	/*
	 * allowed values are positive value or constant for unknown start
	 * index
	 */
	if rangeBasedQuery.StartIndex < 0 {

		return &exception.RVSError{
			Errordetails: "Start Index can not be negative for range based simulation query",
			Errorcode:    "11002",
			Errortype:    "TYPE_QUERY_FAILED",
		}
	}
	/*
	 * allowed values are positive value or constant for unknown end
	 * index
	 */
	if rangeBasedQuery.EndIndex == 0 || rangeBasedQuery.EndIndex < -1 {

		return &exception.RVSError{
			Errordetails: "End Index can not be negative for range based simulation query",
			Errorcode:    "11002",
			Errortype:    "TYPE_QUERY_FAILED",
		}

	}
	/*
	 * Step can have only positive value;
	 */
	if rangeBasedQuery.Step < 1 {
		return &exception.RVSError{
			Errordetails: "Step can not be zero or negative for range based simulation query",
			Errorcode:    "11002",
			Errortype:    "TYPE_QUERY_FAILED",
		}
	}

	var countBasedQuery = SimulationQuery.SimulationCountBasedQuery

	/*
	 * allowed values are positive value or constant for unknown start
	 * and end index
	 */
	if countBasedQuery.StartIndex < -1 {

		return &exception.RVSError{
			Errordetails: "Start Index can not be negative for count based simulation query",
			Errorcode:    "11002",
			Errortype:    "TYPE_QUERY_FAILED",
		}
	}
	if countBasedQuery.Step == 0 {

		return &exception.RVSError{
			Errordetails: "Step can not be zero for count based simulation query",
			Errorcode:    "11002",
			Errortype:    "TYPE_QUERY_FAILED",
		}

	}
	if countBasedQuery.Count < 1 {

		return &exception.RVSError{
			Errordetails: "Count can not be zero or negative for count based simulation query",
			Errorcode:    "11002",
			Errortype:    "TYPE_QUERY_FAILED",
		}
	}
	return nil

}

func createTmpOutputFile(username string, password string, sFileName string) string {

	var outputFileFolder = common.AllocateUniqueFolder(
		common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+common.RM_TOC_XML_FILES, "RVP_GRAPH")
	var outputFile = common.AllocateFile(sFileName, outputFileFolder, username, password)
	return outputFile
}

func getUniqueColumnNamesList(lstColumnNames []string) []string {

	keys := make(map[string]bool)
	uniqueColumnNamesSet := []string{}

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range lstColumnNames {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			uniqueColumnNamesSet = append(uniqueColumnNamesSet, entry)
		}
	}
	return uniqueColumnNamesSet

}

func runPlotDataExtractor(rvpGetPlotModel RVPProcessDataModel) Res {

	var rvpPlotDataQuery datamodel.RVPPlotDataQueryCType

	jsonFile, err := os.Open(rvpGetPlotModel.RequestFilePath)
	// // if we os.Open returns an error then handle it
	if err != nil {
		log.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &rvpPlotDataQuery)

	var rvpPlotDataModel RVPPlotDataModel
	rvpPlotDataModel.MapColumnPoints = make(map[string][]string)
	rvpPlotDataModel.PlotName = rvpPlotDataQuery.RVPPlotColumnInfo.PlotName

	var setUniqueColumnNames = getUniqueColumnNamesList(rvpPlotDataQuery.RVPPlotColumnInfo.ColumnNames)
	var lstColumnPoints []string
	for _, columnName := range setUniqueColumnNames {
		rvpPlotDataModel.MapColumnPoints[columnName] = lstColumnPoints
	}

	var tempSimulationQuery = buildSimulationTempQuery(rvpPlotDataQuery.RVPSimulationQuery)

	if strings.HasSuffix(rvpGetPlotModel.RvpFileExtension, common.RVP_FILE_EXTENSION) {
		rvpPlotDataModel, _ = RVPFilePlotDataExtractor(rvpGetPlotModel.RvpResultFilePath, rvpGetPlotModel, rvpPlotDataModel, tempSimulationQuery)

	} else {
		rvpPlotDataModel = GenericFilePlotDataExtractor(rvpGetPlotModel.RvpResultFilePath, rvpGetPlotModel, rvpPlotDataModel, tempSimulationQuery)
	}

	return buildResponses(rvpPlotDataModel, rvpPlotDataQuery)

}

func buildSimulationTempQuery(RVPSimulationQueryCType *datamodel.RvpsimulationQuery) TemporarySimulationQuery {

	var tempSimulationQuery TemporarySimulationQuery
	var startIndex = 0
	var endIndex = 0
	var step = 0
	var count = 0

	if RVPSimulationQueryCType.RVPSimulationRangeBasedQuery != nil {
		var tempRangeBasedSimulationQuery = RVPSimulationQueryCType.RVPSimulationRangeBasedQuery
		if tempRangeBasedSimulationQuery.StartIndex == common.SIMULATION_START_INDEX {
			tempSimulationQuery.StartIndex = 1
		} else {
			tempSimulationQuery.StartIndex = tempRangeBasedSimulationQuery.StartIndex
		}
		tempSimulationQuery.EndIndex = tempRangeBasedSimulationQuery.EndIndex
		tempSimulationQuery.Step = tempRangeBasedSimulationQuery.Step
		tempSimulationQuery.Count = count
	} else if RVPSimulationQueryCType.RVPSimulationCountBasedQuery != nil {
		var simulationCountBasedQuery = RVPSimulationQueryCType.RVPSimulationCountBasedQuery

		if simulationCountBasedQuery.StartIndex == common.SIMULATION_START_INDEX {
			simulationCountBasedQuery.StartIndex = 1
		}

		if simulationCountBasedQuery.StartIndex != common.SIMULATION_END_INDEX {
			if simulationCountBasedQuery.Step < 0 {
				endIndex = simulationCountBasedQuery.StartIndex
				startIndex = endIndex + simulationCountBasedQuery.Step*(simulationCountBasedQuery.Count-1)
				// Start Index for simulation can not be negative
				if startIndex <= 0 {
					startIndex = 1
				}
				// make step value positive
				step = -(simulationCountBasedQuery.Step)
			} else if simulationCountBasedQuery.Step > 0 {
				startIndex = simulationCountBasedQuery.StartIndex
				endIndex = startIndex + simulationCountBasedQuery.Step*(simulationCountBasedQuery.Count-1)
				step = simulationCountBasedQuery.Step
			}
		} else {
			endIndex = simulationCountBasedQuery.StartIndex
			startIndex = 1
			/*
			 * Step should be negative in case its last n record kind of
			 * query
			 */
			step = -(simulationCountBasedQuery.Step)
			count = simulationCountBasedQuery.Count
		}
		tempSimulationQuery.StartIndex = startIndex
		tempSimulationQuery.EndIndex = endIndex
		tempSimulationQuery.Step = step
		tempSimulationQuery.Count = count
	}
	// System.out.println("Completed buildSimulationTempQuery)()........................");
	return tempSimulationQuery

}

func buildResponses(rvpPlotDataModel RVPPlotDataModel, rvpPlotDataQuery datamodel.RVPPlotDataQueryCType) Res {
	var resData Res

	var lstCurveNames = rvpPlotDataQuery.RVPPlotColumnInfo.ColumnNames
	var Response response
	for i := 0; i < len(lstCurveNames); i++ {
		var datapoints = rvpPlotDataModel.MapColumnPoints[lstCurveNames[i]]
		var dataPointsfloat []float64
		for i := 0; i < len(datapoints); i++ {
			if n, err := strconv.ParseFloat(datapoints[i], 64); err == nil {
				dataPointsfloat = append(dataPointsfloat, n)
			}
		}

		Response.ResponseData.DataSource.Items = dataPointsfloat
		Response.ResponseData.DataSource.NoOfItems = len(rvpPlotDataModel.MapColumnPoints[lstCurveNames[i]])
		Response.ResponseData.DataSource.IsRowVector = true
		resData.Responses.Responselist = append(resData.Responses.Responselist, Response)
	}

	return resData

}
