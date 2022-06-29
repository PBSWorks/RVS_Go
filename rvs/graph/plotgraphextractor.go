package graph

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"altair/rvs/toc"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const VECTOR_NAME = "V"
const MATRIX_NAME = "M"
const SIMULATION_START_INDEX = 0
const SIMULATION_END_INDEX = -1
const DATA_DIR_NAME = "data"

var plotPointValueDecimalPrecision = 8
var ORIGINAL_PLOT_DATA_CSV_FILE_NAME = "/Data_Original.csv"
var PLOT_DATA_CSV_FILE_NAME = "/Data.csv"

const TEMPORARY_PLT_FILE_NAME = "Untitled.plt"

/**
 * Constant for max h3d file size preference
 */
const USER_DATA_POINT_SIZE = "10"

/**
 * Constant for max h3d file size preference
 */
const USER_LINE_THICKNESS = "4"

/**
 * Constant for max h3d file size preference
 */
const ENABLE_DATA_POINT = false

/**
 * Constant for max h3d file size preference
 */
const CURVE_DEFAULT_COLOR = "red"

var matchingFileList MatchingFiles

func GetPlotGraphExtractor(plotRequestResModel PlotRequestResModel, plotRequestCaller string, username string,
	password string, token string) string {

	var indexValue = common.GetUniqueRandomIntValue()
	var plotQueries = buildPlotQueries(plotRequestResModel, token, indexValue)
	plotRequestResModel.PlotRequestResponseModel.Queries = plotQueries
	var resData Res
	if isRVPPlotQuery(plotRequestResModel.PlotRequestResponseModel.Queries.Query[0]) {
		resData = getRVPPlot(plotQueries, plotRequestResModel.ResultFileInformationModel, username, password)
	} else {

		resData = getNativePlot(plotQueries.Query, plotQueries.ResultDataSource[0], plotRequestResModel.ResultFileInformationModel,
			username, password)
	}

	var responses = createResposes(resData.Responses, plotRequestResModel.PlotRequestResponseModel.Queries.Query, "")

	var lstcmPlotModel []plotRequestResponseModel
	var lstPlotModel []plotRequestResponseModel
	lstcmPlotModel = append(lstcmPlotModel, plotRequestResModel.PlotRequestResponseModel)

	for i := 0; i < len(lstcmPlotModel); i++ {
		lstcmPlotModel[0].Responses = extractedResoresponses(responses, lstcmPlotModel[0].Queries)
		lstPlotModel = append(lstPlotModel, lstcmPlotModel[0])
	}
	var plotRequestResModeloutput = CreatePlotResponseModel(plotRequestResModel.ResultFileInformationModel, lstPlotModel,
		getDataDirectoryPath(plotRequestResModel.ResultFileInformationModel.ServerName, username), len(lstPlotModel), token, plotRequestCaller)

	if (plotRequestCaller == "FROM_TOC") ||
		(plotRequestResModel.PlotRequestResponseModel.PlotMetaData.UserPreferece.UserPrefereces[0].Name == "") ||
		(len(plotRequestResModel.PlotRequestResponseModel.PlotMetaData.UserPreferece.UserPrefereces) == 0) {
		plotRequestResModeloutput.PlotRequestResponseModel.PlotMetaData.UserPreferece = GetUserPlotPreferences()
	}

	var matchingFileListData = GetWLMFileList(plotRequestResModel.ResultFileInformationModel, token,
		plotRequestResModel.ResultFileInformationModel.FilePath, plotRequestResModel.ResultFileInformationModel)
	plotRequestResModeloutput.LstMatchingResultFiles = matchingFileListData

	if xmlstring, err := json.MarshalIndent(plotRequestResModeloutput, "", "    "); err == nil {
		return string(xmlstring)
	}

	return ""
}

func isRVPPlotQuery(singlequery Query) bool {
	if singlequery.RvpPlotDataQuery.RvpPlotColumnInfo.ColumnName != "" {
		return true
	} else {
		return false
	}
}

func getNativePlot(lstOfQuesries []Query, datasource datamodel.ResourceDataSource,
	ResultFileInformationModel resultFileInformationModel, username string, password string) Res {
	if len(lstOfQuesries) == 0 {
		log.Println("No query present in the request")
		res := Res{}
		return res
	}
	var resData Res
	var tempOmlFile = createTempOMLFile(username, password)
	var dataOutputFile = createDataOutputFile(username, password)
	var sMasterOmlFileName = createMasterOMLFile()

	var sJobId = ResultFileInformationModel.JobId
	var sJobState = ResultFileInformationModel.JobState
	var sResultFilePath = ""
	if sJobId == "" && sJobState == "" {
		sResultFilePath, _ = common.ResolveFilePortDataSource(datasource, username, password)
		// if resulrdatasourceerr != nil {
		// 	return "", resulrdatasourceerr
		// }
	} else {
		sResultFilePath, _ = common.ResolvePBSPortDataSource(datasource, username, password)
		// if resulrdatasourceerr != nil {
		// 	return "", resulrdatasourceerr
		// }
	}

	writeIntoOmlFile(lstOfQuesries, tempOmlFile, dataOutputFile, lstOfQuesries[0].ResultDataSourceRef[0].Id, sResultFilePath, sMasterOmlFileName)

	toc.ExecuteComposeApplicatopn(tempOmlFile, username, password)
	jsonFile, err := os.Open(dataOutputFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &resData)

	return resData

}

func extractedResoresponses(resData responses, Queries queries) responses {
	var numOfQueries = len(Queries.Query)
	var responsesForQueries responses
	// /var lstResponseCTypes []response
	for i := 0; i < numOfQueries; i++ {
		var responseCType = searchResponse(resData, Queries.Query[i].VarName)
		// /lstResponseCTypes = append(lstResponseCTypes, responseCType)
		//responses.getResponse().remove(responseCType);
		responsesForQueries.Responselist = append(responsesForQueries.Responselist, responseCType)
	}
	return responsesForQueries
}

func searchResponse(resData responses, queryid string) response {
	var returnResponseCType response
	for i := 0; i < len(resData.Responselist); i++ {
		if resData.Responselist[i].Id == queryid {
			returnResponseCType = resData.Responselist[i]
			break
		}
	}
	return returnResponseCType
}

func createTempOMLFile(username string, password string) string {
	var tempOmlFolder = common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+common.RM_SCRIPT_FILES, "PLOT_GRAPH")
	var tempOmlFile = common.AllocateFile(common.TEMP_OML_FILE_NAME, tempOmlFolder, username, password)
	return tempOmlFile

}

func createDataOutputFile(username string, password string) string {
	var dataOutputFolder = common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+common.RM_OUTPUT_FILES, "PLOT_GRAPH")
	var dataOutputFile = common.AllocateFile("RawData.json", dataOutputFolder, username, password)
	return dataOutputFile

}

func createMasterOMLFile() string {

	var sMasterOmlFileName = common.GetRSHome() + common.Graph_MASTER_OML_FILE_NAME

	sMasterOmlFileName = strings.Replace(sMasterOmlFileName, common.BACK_SLASH, common.FORWARD_SLASH, -1)
	return sMasterOmlFileName
}

func writeIntoOmlFile(lstOfQueries []Query, tempOmlFile string, dataOutputFile string, datasourceid string, sResultFilePath string, sMasterOmlFileName string) {

	// Output file declaration
	var sOutputFileName = strings.Replace(dataOutputFile, common.BACK_SLASH, common.FORWARD_SLASH, -1)

	var firstline = "global HWEP_RAWDATA_OUTPUTFILE;" + common.NEWLINE + "HWEP_RAWDATA_OUTPUTFILE = " +
		common.SINGLE_QUOTE + sOutputFileName + common.SINGLE_QUOTE + common.NEWLINE + ";"

	var secondline = "global " + datasourceid + ";" + common.NEWLINE + datasourceid + " = " +
		common.SINGLE_QUOTE + sResultFilePath + common.SINGLE_QUOTE + ";" + common.NEWLINE

	var thirdline = "status = addpath (" + common.SINGLE_QUOTE + sMasterOmlFileName + common.SINGLE_QUOTE + ");"

	var forthline = "run (" + common.SINGLE_QUOTE + common.PLOT_GRAPH_OML_FILE_NAME + common.SINGLE_QUOTE + ");"

	var fifthline = "GET_RAWDATA_HEADER();" + common.NEWLINE

	plotgraphfilecontent := []string{firstline, secondline, thirdline, forthline, fifthline}

	plotgraphfilecontent = append(plotgraphfilecontent, "GET_RAWDATA_RESPONSE_HEADER();"+common.NEWLINE)
	for i := 0; i < len(lstOfQueries); i++ {

		//var sResFileVarName = lstOfQueries[i].ResultDataSourceRef[0].Id
		var sResFileVarName = datasourceid
		var sVarName string
		if lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Type.Name != "" {
			if lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name != "" {
				if (lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Subcase.Index < 0) ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Type.Name == "") ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name == "") ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.Component.Name == "") {
					log.Println("Query failed due to invalid data")
					// throw new RMFrameworkException(RMFrameworkException.CODE_MISSING_QUERY_DATA,
					// 	RMFrameworkException.TYPE_QUERY_FAILED);
				}

				//Check if variable name is provided in query. If yes, use it
				if common.IsValidString(lstOfQueries[i].VarName) {
					sVarName = strings.TrimSpace(lstOfQueries[i].VarName)
				} else {
					sVarName = common.SINGLE_QUOTE + VECTOR_NAME + common.SINGLE_QUOTE
				}

				arrArguements := [6]string{}
				if lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Subcase.Index >= 1 {
					if !common.IsValidString(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Subcase.Name) {
						log.Println("Invalid data: Subcase name missing in the query")
						// throw new RMFrameworkException(RMFrameworkException.CODE_INVALID_QUERY_DATA,
						// 				RMFrameworkException.TYPE_QUERY_FAILED);
					}

					arrArguements[0] = sResFileVarName
					arrArguements[1] = lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Subcase.Name
					arrArguements[2] = lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Type.Name
					arrArguements[3] = lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name
					arrArguements[4] = lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.Component.Name
					arrArguements[5] = sVarName
					arrCompArgs := [6]string{}
					//arrCompArgs[0] = mapResfileIdVsPath.get(sResFileVarName);
					arrCompArgs[0] = sResultFilePath
					arrCompArgs[1] = common.GetPlatformIndependentFilePath(common.GetRSHome()+common.Graph_MASTER_COMPONENT_FILE_PATH, false) + common.COMP_LIST_FILE_PATH
					arrCompArgs[2] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Subcase.Index)
					arrCompArgs[3] = arrArguements[2]
					arrCompArgs[4] = arrArguements[3]
					arrCompArgs[5] = arrArguements[4]

					plotgraphfilecontent = append(plotgraphfilecontent, "[a,subcase,datatype,request,component]=doesComponentExist("+
						common.SINGLE_QUOTE+arrCompArgs[0]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+arrCompArgs[1]+common.SINGLE_QUOTE+","+
						arrCompArgs[2]+","+
						common.SINGLE_QUOTE+arrCompArgs[3]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+arrCompArgs[4]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+arrCompArgs[5]+common.SINGLE_QUOTE+")"+
						common.NEWLINE+"if (a==1)"+common.NEWLINE)

					plotgraphfilecontent = append(plotgraphfilecontent,
						arrArguements[5]+" = readvector ("+
							arrArguements[0]+","+
							common.SINGLE_QUOTE+arrArguements[1]+common.SINGLE_QUOTE+","+
							common.SINGLE_QUOTE+arrArguements[2]+common.SINGLE_QUOTE+","+
							common.SINGLE_QUOTE+arrArguements[3]+common.SINGLE_QUOTE+","+
							common.SINGLE_QUOTE+arrArguements[4]+common.SINGLE_QUOTE+");"+common.NEWLINE)
				} else {
					arrArguements := [5]string{}
					arrArguements[0] = sResFileVarName
					arrArguements[1] = lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Type.Name
					arrArguements[2] = lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name
					arrArguements[3] = lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.Component.Name
					arrArguements[4] = sVarName

					arrCompArgs := [6]string{}
					//arrCompArgs[0] = mapResfileIdVsPath.get(sResFileVarName);
					arrCompArgs[0] = sResultFilePath
					arrCompArgs[1] = common.GetPlatformIndependentFilePath(common.GetRSHome()+common.Graph_MASTER_COMPONENT_FILE_PATH, false) + common.COMP_LIST_FILE_PATH
					arrCompArgs[2] = strconv.Itoa(1)
					arrCompArgs[3] = arrArguements[1]
					arrCompArgs[4] = arrArguements[2]
					arrCompArgs[5] = arrArguements[3]

					plotgraphfilecontent = append(plotgraphfilecontent, "[b,subcase1,datatype1,request1,component1]=doesComponentExist("+
						common.SINGLE_QUOTE+arrCompArgs[0]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+arrCompArgs[1]+common.SINGLE_QUOTE+","+
						arrCompArgs[2]+","+
						common.SINGLE_QUOTE+arrCompArgs[3]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+arrCompArgs[4]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+arrCompArgs[5]+common.SINGLE_QUOTE+");"+
						common.NEWLINE+"if (b==1)"+common.NEWLINE)

					plotgraphfilecontent = append(plotgraphfilecontent,
						arrArguements[4]+" = readvector ("+
							arrArguements[0]+","+
							common.SINGLE_QUOTE+arrArguements[1]+common.SINGLE_QUOTE+","+
							common.SINGLE_QUOTE+arrArguements[2]+common.SINGLE_QUOTE+","+
							common.SINGLE_QUOTE+arrArguements[3]+common.SINGLE_QUOTE+","+
							common.SINGLE_QUOTE+arrArguements[4]+common.SINGLE_QUOTE+");"+common.NEWLINE)
				}
				if lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Start != 0 {
					subsetArguements := [5]string{}
					subsetArguements[0] = sVarName
					subsetArguements[1] = sVarName
					subsetArguements[2] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Start)
					subsetArguements[3] = strconv.FormatInt(lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.End, 10)
					if lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Step != 0 {
						subsetArguements[4] = strconv.FormatInt(lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Step, 10)
					} else {
						subsetArguements[4] = strconv.Itoa(1)
					}

					plotgraphfilecontent = append(plotgraphfilecontent, subsetArguements[0]+" = GET_VECTOR_SUBSET ("+
						common.SINGLE_QUOTE+subsetArguements[1]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+subsetArguements[2]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+subsetArguements[3]+common.SINGLE_QUOTE+","+
						common.SINGLE_QUOTE+subsetArguements[4]+common.SINGLE_QUOTE+");"+common.NEWLINE)
				}

				if lstOfQueries[i].PlotResultQuery.DataQuery.SimulationQuery.SimulationRangeBasedQuery.StartIndex == 0 {
					plotgraphfilecontent = append(plotgraphfilecontent,
						getSimulationQuery(sVarName, lstOfQueries[i].PlotResultQuery.DataQuery.SimulationQuery))
				}
				if lstOfQueries[i].PlotResultQuery.DataQuery.IsRawDataRequired {
					plotgraphfilecontent = append(plotgraphfilecontent, "HWEP_RS_OUTPUT_VAR("+common.SINGLE_QUOTE+"VECTOR"+common.SINGLE_QUOTE+","+
						sVarName+");"+common.NEWLINE)
					if i+1 != len(lstOfQueries) {
						plotgraphfilecontent = append(plotgraphfilecontent, "fidXMLOutput = fopen(HWEP_RAWDATA_OUTPUTFILE,'a');"+common.NEWLINE)
						plotgraphfilecontent = append(plotgraphfilecontent, "fwrite(fidXMLOutput,',');"+common.NEWLINE)
						plotgraphfilecontent = append(plotgraphfilecontent, "fclose(fidXMLOutput);"+common.NEWLINE)
					}
				}
			} else if lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.DataRequestIndex.Start != 0 {
				if (lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Type.Name == "") ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Type.Index <= 0) ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.DataRequestIndex.Start <= 0) ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.DataRequestIndex.End <= 0) ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.ComponentIndex.Start <= 0) ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.ComponentIndex.End <= 0) ||
					(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.TimeStep.Index <= 0) {
					log.Println("Query failed due to invalid data")
					// LOGGER.error(sMessage);
					// throw new RMFrameworkException(RMFrameworkException.CODE_INVALID_QUERY_DATA,
					// 				RMFrameworkException.TYPE_QUERY_FAILED);
				}
				arrArguements := [7]string{}
				arrArguements[0] = sResFileVarName
				arrArguements[1] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.Type.Index)
				arrArguements[2] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.DataRequestIndex.Start)

				arrArguements[3] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.DataRequestIndex.End)
				arrArguements[4] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.ComponentIndex.Start)
				arrArguements[5] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.ComponentIndex.End)
				arrArguements[6] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.StrcQuery.ContiguousRequest.TimeStep.Index)

				plotgraphfilecontent = append(plotgraphfilecontent, "M = ReadVectors ("+
					common.SINGLE_QUOTE+arrArguements[0]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+arrArguements[1]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+arrArguements[2]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+arrArguements[3]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+arrArguements[4]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+arrArguements[5]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+arrArguements[6]+common.SINGLE_QUOTE+");"+common.NEWLINE)

				if lstOfQueries[i].PlotResultQuery.DataQuery.IsRawDataRequired {
					plotgraphfilecontent = append(plotgraphfilecontent, "HWEP_RS_OUTPUT_VAR(\"MATRIX\", M);"+common.NEWLINE)
					if i+1 != len(lstOfQueries) {
						plotgraphfilecontent = append(plotgraphfilecontent, "fidXMLOutput = fopen(HWEP_RAWDATA_OUTPUTFILE,'a');"+common.NEWLINE)
						plotgraphfilecontent = append(plotgraphfilecontent, "fwrite(fidXMLOutput,',');"+common.NEWLINE)
						plotgraphfilecontent = append(plotgraphfilecontent, "fclose(fidXMLOutput);"+common.NEWLINE)
					}
				}
				//Check if variable name is provided in query. If yes, use it
				// if common.IsValidString(lstOfQueries[i].VarName) {
				// 	sVarName = lstOfQueries[i].VarName
				// } else {
				// 	sVarName = MATRIX_NAME
				// }
			} else {
				log.Println("No query found")
				// LOGGER.error(sMessage);
				// throw new RMFrameworkException(RMFrameworkException.CODE_MISSING_QUERY_DATA,
				// 				RMFrameworkException.TYPE_QUERY_FAILED);
			}
			//getStatsInfo(strcQuery, omlFileWriter, sVarName)
			//getSamplingInfo(strcQuery, omlFileWriter, sVarName)
			plotgraphfilecontent = append(plotgraphfilecontent, "else"+common.NEWLINE)
			plotgraphfilecontent = append(plotgraphfilecontent, "writeComponentError();"+common.NEWLINE)
			plotgraphfilecontent = append(plotgraphfilecontent, "end"+common.NEWLINE)
		} else if lstOfQueries[i].PlotResultQuery.DataQuery.InlineQuery.Expression != "" {
			/* if any uncached query */
			var sExpresssion = lstOfQueries[i].PlotResultQuery.DataQuery.InlineQuery.Expression

			//List<String> queriesInExpression = new ArrayList<String>();

			// for( ResponseCType  responseCType : lstResponseCType)
			// {
			// 	if( sExpresssion.indexOf(responseCType.getId()) != -1)
			// 	{
			// 		omlFileWriter.write(responseCType.getId() + " = " + convertResonseToString(responseCType, maxlength)+ ICommonConstants.NEWLINE);
			// 	}
			// }

			sVarName = lstOfQueries[i].VarName

			if common.IsValidString(sExpresssion) {
				var iIndex = strings.Index(sExpresssion, "=")
				if (iIndex < 0) && common.IsValidString(sVarName) {
					plotgraphfilecontent = append(plotgraphfilecontent, sVarName+" = "+sExpresssion+common.NEWLINE)
				} else {
					sVarName = sExpresssion[0:iIndex]
					plotgraphfilecontent = append(plotgraphfilecontent, sExpresssion+common.NEWLINE)
				}
			}
			if lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Start != 0 {
				subsetArguements := [5]string{}
				subsetArguements[0] = sVarName
				subsetArguements[1] = sVarName
				subsetArguements[2] = strconv.Itoa(lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Start)
				subsetArguements[3] = strconv.FormatInt(lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.End, 10)
				if lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Step != 0 {
					subsetArguements[4] = strconv.FormatInt(lstOfQueries[i].PlotResultQuery.DataQuery.SimulationFilter.Step, 10)
				} else {
					subsetArguements[4] = strconv.Itoa(1)
				}

				plotgraphfilecontent = append(plotgraphfilecontent, subsetArguements[0]+" = GET_VECTOR_SUBSET ("+
					common.SINGLE_QUOTE+subsetArguements[1]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+subsetArguements[2]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+subsetArguements[3]+common.SINGLE_QUOTE+","+
					common.SINGLE_QUOTE+subsetArguements[4]+common.SINGLE_QUOTE+");"+
					common.NEWLINE)
			}
			if lstOfQueries[i].PlotResultQuery.DataQuery.IsRawDataRequired {
				plotgraphfilecontent = append(plotgraphfilecontent, "HWEP_RS_OUTPUT_VAR(\"VECTOR\","+sVarName+");"+common.NEWLINE)
				if i+1 != len(lstOfQueries) {
					plotgraphfilecontent = append(plotgraphfilecontent, "fidXMLOutput = fopen(HWEP_RAWDATA_OUTPUTFILE,'a');"+common.NEWLINE)
					plotgraphfilecontent = append(plotgraphfilecontent, "fwrite(fidXMLOutput,',');"+common.NEWLINE)
					plotgraphfilecontent = append(plotgraphfilecontent, "fclose(fidXMLOutput);"+common.NEWLINE)
				}
			}
		} else {
			log.Println("No query found")
			// LOGGER.error(sMessage);
			// throw new RMFrameworkException(RMFrameworkException.CODE_MISSING_QUERY_DATA,
			// 				RMFrameworkException.TYPE_QUERY_FAILED);
		}
	}
	plotgraphfilecontent = append(plotgraphfilecontent, "GET_RAWDATA_RESPONSE_FOOTER();"+common.NEWLINE)
	plotgraphfilecontent = append(plotgraphfilecontent, "GET_RAWDATA_FOOTER();"+common.NEWLINE)

	file, err := os.OpenFile(tempOmlFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	datawriter := bufio.NewWriter(file)

	for _, data := range plotgraphfilecontent {
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	file.Close()

}

func getSimulationQuery(sVarName string, SimulationQuery simulationQuery) string {

	var startIndex string
	var endIndex string
	var step string
	if SimulationQuery.SimulationCountBasedQuery.StartIndex == 0 &&
		SimulationQuery.SimulationCountBasedQuery.Count == 0 &&
		SimulationQuery.SimulationCountBasedQuery.Step == 0 {
		if SimulationQuery.SimulationRangeBasedQuery.StartIndex == SIMULATION_START_INDEX {
			startIndex = "1"
		} else {
			startIndex = strconv.Itoa(SimulationQuery.SimulationRangeBasedQuery.StartIndex)
		}
		if SimulationQuery.SimulationRangeBasedQuery.EndIndex == SIMULATION_END_INDEX {
			endIndex = "length(" + sVarName + ")"
		} else {
			endIndex = strconv.Itoa(SimulationQuery.SimulationRangeBasedQuery.EndIndex)
		}
		step = strconv.Itoa(SimulationQuery.SimulationRangeBasedQuery.Step)
	} else if SimulationQuery.SimulationRangeBasedQuery.StartIndex == 0 &&
		SimulationQuery.SimulationRangeBasedQuery.EndIndex == 0 &&
		SimulationQuery.SimulationRangeBasedQuery.Step != 0 {
		// last n based query
		if SimulationQuery.SimulationCountBasedQuery.Step < 0 {
			if SimulationQuery.SimulationCountBasedQuery.StartIndex == SIMULATION_END_INDEX {
				endIndex = "length(" + sVarName + ")"
			} else {
				endIndex = strconv.Itoa(SimulationQuery.SimulationCountBasedQuery.StartIndex)
			}

			startIndex = endIndex + "+" + "(" +
				strconv.Itoa(SimulationQuery.SimulationCountBasedQuery.Count) + " - 1)*" +
				strconv.Itoa(SimulationQuery.SimulationCountBasedQuery.Step)
			step = strconv.Itoa(-SimulationQuery.SimulationCountBasedQuery.Step)
		} else {
			if SimulationQuery.SimulationCountBasedQuery.StartIndex == SIMULATION_START_INDEX {
				startIndex = "1"
			} else {
				startIndex = strconv.Itoa(SimulationQuery.SimulationCountBasedQuery.StartIndex)
			}
			endIndex = startIndex + "+" + "(" +
				strconv.Itoa(SimulationQuery.SimulationCountBasedQuery.Count) + " - 1)*" +
				strconv.Itoa(SimulationQuery.SimulationCountBasedQuery.Step)
			step = strconv.Itoa(SimulationQuery.SimulationCountBasedQuery.Step)
		}
	}
	subsetArguements := [5]string{}
	subsetArguements[0] = sVarName
	subsetArguements[1] = sVarName
	subsetArguements[2] = startIndex
	subsetArguements[3] = endIndex
	subsetArguements[4] = step

	return subsetArguements[0] + " = GET_VECTOR_SUBSET (" +
		subsetArguements[1] + "," +
		subsetArguements[2] + "," +
		subsetArguements[3] + "," +
		subsetArguements[4] + ");" + common.NEWLINE

}

func CreatePlotResponseModel(ResultFileInformationModel resultFileInformationModel, lstPlotModel []plotRequestResponseModel,
	sDataDirectoryPath string, newlyAddedPltBlocksCount int, token string, plotRequestCaller string) PlotRequestResModel {

	var plotRequestResModel PlotRequestResModel
	plotRequestResModel.ResultFileInformationModel = ResultFileInformationModel

	var plotAmChartsdata plotAmCharts
	plotAmChartsdata.ChartHtmlRelativeUrl = "/ui/cm/plugins/rm/data/Chart.html"
	plotAmChartsdata.ExportPlotDataRelativeUrl = "/resultmanagerservice/rest/rmservice/exportplotdata/"
	plotAmChartsdata.PlotFileRelativePath = getDataDirectoryRelativePath(sDataDirectoryPath)

	var mergedPlotTemporaryModel = getMergedPlotTemporaryModel(ResultFileInformationModel.FileName, lstPlotModel)

	var plotAmChartsdataupdated = buildPlotAMChartsModel(mergedPlotTemporaryModel, sDataDirectoryPath, plotAmChartsdata)

	var tempFilePath = createTemporaryPLTFile(sDataDirectoryPath, lstPlotModel)

	/*added for plot comparision **/
	for i := 0; i < len(lstPlotModel); i++ {
		plotRequestResModel.PlotRequestResponseModel.Queries.Query =
			append(plotRequestResModel.PlotRequestResponseModel.Queries.Query, lstPlotModel[i].Queries.Query...)
	}

	plotRequestResModel.PlotRequestResponseModel.PlotResponseModel.PlotAmCharts = plotAmChartsdataupdated
	plotRequestResModel.PlotRequestResponseModel.PlotResponseModel.TemporaryPltFilePath = tempFilePath
	plotRequestResModel.PlotRequestResponseModel.PlotResponseModel.NewlyAddedPltBlocksCount = newlyAddedPltBlocksCount
	plotRequestResModel.PlotRequestResponseModel.PlotMetaData.TitleMetaData = lstPlotModel[0].PlotMetaData.TitleMetaData
	plotRequestResModel.PlotRequestResponseModel.PlotMetaData.GraphMetaData = lstPlotModel[0].PlotMetaData.GraphMetaData

	return plotRequestResModel

}

func getDataDirectoryPath(servername string, username string) string {
	var sDataDirectoryPath = getNewDirPath(servername, username)
	if err := os.MkdirAll(sDataDirectoryPath, 0755); err != nil {
		log.Fatal(err)
	}

	return sDataDirectoryPath
}

func getNewDirPath(sServerName string, username string) string {
	var sUniqueDirPath string = ""
	sUniqueDirPath = GetRMDataDirectory() + username + "/" + strconv.FormatInt(common.GetUniqueRandomIntValue(), 10)
	//log.Fatal("Error occured while creating new directory path")
	return sUniqueDirPath
}

func getDataDirectoryRelativePath(sDataDirectoryPath string) string {
	return sDataDirectoryPath[len([]rune(GetRMDataDirectory())):len(sDataDirectoryPath)]
}

func GetRMDataDirectory() string {
	//return os.Getenv("PBSWORKS_HOME") + "/data/resultmanager" + "/" + DATA_DIR_NAME + "/"

	// var dir = common.GetRSHome() + "/data/"
	// if err := os.Mkdir(dir, 0755); err != nil {
	// 	log.Fatal(err)
	// }
	return common.GetRSHome() + "/data/"

}

func getMergedPlotTemporaryModel(ResultFileName string, plotRequestResponseModelList []plotRequestResponseModel) PlotTemporaryModel {

	var pltModel PlotTemporaryModel
	var lstPlotTemporaryModelList []PlotTemporaryModel
	for i := 0; i < len(plotRequestResponseModelList); i++ {
		if isRVPPlotQuery(plotRequestResponseModelList[i].Queries.Query[0]) {
			pltModel = readRVPPLTModel(ResultFileName, plotRequestResponseModelList[i])
		} else {
			pltModel = readNativePLTModel(ResultFileName, plotRequestResponseModelList[i])
		}
		lstPlotTemporaryModelList = append(lstPlotTemporaryModelList, pltModel)
	}
	var mergedPlotTemporaryModel PlotTemporaryModel

	var plotTemporaryModelListSize = len(plotRequestResponseModelList)
	/*
	 * In case of only one plot temporary model, that becomes
	 * the merged result
	 */
	mergedPlotTemporaryModel = lstPlotTemporaryModelList[0]
	if plotTemporaryModelListSize > 1 {
		for i := 1; i < plotTemporaryModelListSize; i++ {
			mergedPlotTemporaryModel = mergeCurveNamesAndPoints(mergedPlotTemporaryModel, lstPlotTemporaryModelList[i])
		}
	}
	/**
	 * Plot meta data information will be taken from the first PlotTemporaryModel
	 * Logic can be changed in future.
	 */
	mergedPlotTemporaryModel.PlotMetaData = lstPlotTemporaryModelList[0].PlotMetaData
	return mergedPlotTemporaryModel
}

func readNativePLTModel(resultFileName string, pltModel plotRequestResponseModel) PlotTemporaryModel {

	var plotMetaData = pltModel.PlotMetaData
	// String resultFileName = getResultFileName((ResultDataSourceLocalFileCType) pltModel
	//                 .getQueries().getResultDataSource().get(0));
	var resultFileId = pltModel.Queries.Query[0].ResultDataSourceRef[0].Id

	// get plot curve names
	var lstQueries = pltModel.Queries.Query
	var lstLegendNames []string
	var lstCurveNames []string

	var curveName string
	//var strcQuery strcQuery
	var plotQuery Query
	var sLegendName string
	var inlineQueryCTye inlineQuery
	for index := 0; index < len(lstQueries); index++ {
		plotQuery = lstQueries[index]

		/*
		 * Dont add the first query as it is for X Axis
		 */
		if index != 0 {

			if plotQuery.PlotResultQuery.DataQuery.StrcQuery.Subcase.Name != "" {
				curveName = resultFileName + " : " + plotQuery.PlotResultQuery.DataQuery.StrcQuery.Subcase.Name + " : " +
					plotQuery.PlotResultQuery.DataQuery.StrcQuery.Type.Name + " : " +
					plotQuery.PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name + " : " +
					plotQuery.PlotResultQuery.DataQuery.StrcQuery.DistantRequest.Component.Name + "(" + plotQuery.VarName + ")" + " :" +
					resultFileId

				sLegendName = plotQuery.PlotResultQuery.DataQuery.StrcQuery.Type.Name + ":" +
					plotQuery.PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name + ":" +
					plotQuery.PlotResultQuery.DataQuery.StrcQuery.DistantRequest.Component.Name + "(" + plotQuery.VarName + ")"

			} else {
				inlineQueryCTye = plotQuery.PlotResultQuery.DataQuery.InlineQuery
				if inlineQueryCTye.Title != "" {
					curveName = resultFileName + ":" + inlineQueryCTye.Title + ":" +
						inlineQueryCTye.Expression + "(" +
						plotQuery.VarName + ")" + " :" + resultFileId
					sLegendName = inlineQueryCTye.Title + "(" + plotQuery.VarName + ")"
				}
			}
			lstCurveNames = append(lstCurveNames, curveName)
			lstLegendNames = append(lstLegendNames, sLegendName)
		}
	}

	var lstListCurvePoints [][]float64

	var lstResponse = pltModel.Responses.Responselist
	for i := 0; i < len(lstResponse); i++ {
		var responseData = lstResponse[i].ResponseData
		var ds = responseData.DataSource
		lstListCurvePoints = append(lstListCurvePoints, ds.Items)
	}

	return PlotTemporaryModel{
		PlotMetaData:   plotMetaData,
		lstCurveNames:  lstCurveNames,
		lstCurvesData:  lstListCurvePoints,
		lstLegendNames: lstLegendNames,
	}
}

func createResposes(response responses, lstQueryCType []Query, sSessionId string) responses {
	if len(response.Responselist) > 0 {

		for i := 0; i < len(response.Responselist); i++ {

			response.Responselist[i].Id = lstQueryCType[i].VarName
			// SessionCType session = new SessionCType();
			// session.setId(sSessionId);
			// responseCType.setSession(session);

			// Check for filtered component case
			if len(response.Responselist) > 0 {
				if response.Responselist[i].ResponseData.DataSource.Type == "DataSourceStringInlineCType" {

					if response.Responselist[i].ResponseData.DataSource.Type == "COMPONENT_NOT_PRESENT" {
						var sMessage = "Component not present for the specified request"
						log.Fatalf(sMessage)
					}
				}
			}
		}
	} else {
		var sMessage = "Error occurred while getting the response"
		log.Fatalf(sMessage)
	}

	return response

}

func buildPlotAMChartsModel(mergedPlotTemporaryModel PlotTemporaryModel, sDataDirectoryPath string, PlotAmCharts plotAmCharts) plotAmCharts {

	var plotAmChartsdata plotAmCharts
	var plotMetaData = mergedPlotTemporaryModel.PlotMetaData.TitleMetaData
	var lstMergedCurvePoints = mergedPlotTemporaryModel.lstCurvesData
	var lstMergedCurveNames = mergedPlotTemporaryModel.lstCurveNames
	createAMChartFiles(plotMetaData, sDataDirectoryPath, lstMergedCurvePoints, lstMergedCurveNames)
	plotAmChartsdata.ExportPlotDataRelativeUrl = PlotAmCharts.ExportPlotDataRelativeUrl + sDataDirectoryPath +
		ORIGINAL_PLOT_DATA_CSV_FILE_NAME
	plotAmChartsdata.PlotFileRelativePath = PlotAmCharts.PlotFileRelativePath + PLOT_DATA_CSV_FILE_NAME
	var plotDataModelData plotDataModel
	plotDataModelData.CurveNames = lstMergedCurveNames
	plotDataModelData.NumberOfCurvePoints = getNumberOfDataPoints()
	plotDataModelData.DataPoints = getDataPoints(false, false)
	plotDataModelData.LogXdataPoints = getDataPointsLogX()
	plotDataModelData.LogYdataPoints = getDataPointsLogY()
	plotDataModelData.LogXlogYdataPoints = getDataPointsLogYandLogX()
	plotDataModelData.XaxisNegative = doesXCurveContainsNegativeValues()
	plotDataModelData.YaxisNegative = doesYCurveContainsNegativeValues()
	plotDataModelData.LegendNames = mergedPlotTemporaryModel.lstLegendNames
	plotAmChartsdata.PlotDataModel = plotDataModelData

	return plotAmChartsdata
}

func createAMChartFiles(plotMetaData titleMetaData, sDataDirectoryPath string, lstMergedCurvePoints [][]float64,
	lstMergedCurveNames []string) {
	createCSVDataFiles(sDataDirectoryPath, lstMergedCurvePoints,
		plotMetaData.XaxisTitle, plotMetaData.YaxisTitle, lstMergedCurveNames,
		plotPointValueDecimalPrecision)
}

func createTemporaryPLTFile(sDataDirectoryPath string, lstPlotModel []plotRequestResponseModel) string {
	_, err := os.Create(sDataDirectoryPath + "/" + TEMPORARY_PLT_FILE_NAME)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	writeToPLTFile(sDataDirectoryPath+"/"+TEMPORARY_PLT_FILE_NAME, lstPlotModel)
	return sDataDirectoryPath + "/" + TEMPORARY_PLT_FILE_NAME
}

func GetUserPlotPreferences() userPreferece {
	var UserPreferece userPreferece

	UserPreferece.UserPrefereces = append(UserPreferece.UserPrefereces, userPrefereces{
		Name:               "user_preferences",
		CurveDatapointSize: USER_DATA_POINT_SIZE,
		CurveLineThickness: USER_LINE_THICKNESS,
		EnableDataPoint:    ENABLE_DATA_POINT,
		CurveColors:        CURVE_DEFAULT_COLOR,
	})

	return UserPreferece
}

func GetWLMFileList(fileInformationModel resultFileInformationModel, sToken string,
	fileDir string, templateFile resultFileInformationModel) []resultFileInformationModel {

	var builFileListURL = buildGetWLMFileListURL(fileInformationModel, sToken, templateFile)
	var postData = buildPostDataFileList(fileInformationModel, templateFile)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", builFileListURL, bytes.NewBufferString(postData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", sToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal([]byte(body), &matchingFileList)
	var listFileInfoModel []resultFileInformationModel
	if len(matchingFileList.Data.Files) != 0 {

		for i := 0; i < len(matchingFileList.Data.Files); i++ {

			listFileInfoModel = append(listFileInfoModel, resultFileInformationModel{
				FileName:   matchingFileList.Data.Files[i].Filename,
				FilePath:   matchingFileList.Data.Files[i].AbsPath,
				SeriesFile: false,
				ServerName: templateFile.ServerName,
				JobId:      templateFile.JobId,
				JobState:   templateFile.JobState,
				Size:       strconv.FormatInt(matchingFileList.Data.Files[i].Size, 10),
				PasUrl:     templateFile.PasUrl,
			})

		}
	}
	return listFileInfoModel

}

func buildGetWLMFileListURL(fileInformationModel resultFileInformationModel, sToken string,
	templateFile resultFileInformationModel) string {

	var fileListUrl = fileInformationModel.PasUrl

	var access_token = "access_token=" + strings.Replace(sToken, "Bearer", "", -1)
	common.GetWLMDetails(access_token, templateFile.ServerName, templateFile.PasUrl)
	if fileInformationModel.JobState == "R" && strings.Contains(fileListUrl, common.PAS_URL_VALUE) {
		fileListUrl = strings.Replace(fileListUrl, common.PAS_URL_VALUE, common.JOB_OPERATION, -1)
	} else {
		fileListUrl = fileListUrl + common.REST_SERVICE_URL
	}
	fileListUrl = fileListUrl + "/files/list"
	return fileListUrl
}

func buildPostDataFileList(fileInformationModel resultFileInformationModel, templateFile resultFileInformationModel) string {

	var jobId = getJobId(templateFile)
	if jobId == "" {
		jobId = "\"" + "\""
	} else {
		jobId = "\"" + jobId + "\""
	}
	var filter = ""
	if fileInformationModel.FileExtension != "" {
		filter = fileInformationModel.FileExtension
	} else {
		filter = fileInformationModel.FileName
	}

	var iIndex = strings.Index(templateFile.FilePath, templateFile.FileName)
	var fileLoation = templateFile.FilePath[0:iIndex]

	var postData = "{\"path\":\"" + fileLoation + "\",\"jobid\":" + jobId + ",\"filter\":\"" + filter + "\"}"

	return postData
}

func getJobId(fileModel resultFileInformationModel) string {
	var sJobId = ""

	if common.JOB_RUNNING_STATE == fileModel.JobState {
		sJobId = fileModel.JobId
	} else {
		log.Println("Job is not in running state, no need of job id")
	}
	return sJobId
}

func buildPlotQueries(cmPlotRequestResponseModel PlotRequestResModel, sToken string, indexValue int64) queries {
	var ResultDataSource = buildResultFileDataSource(cmPlotRequestResponseModel, sToken, indexValue)
	var updatedqueries = attachDatasourcesToQueries(cmPlotRequestResponseModel.PlotRequestResponseModel.Queries, ResultDataSource)
	cmPlotRequestResponseModel.PlotRequestResponseModel.Queries = updatedqueries
	return updatedqueries
}

func buildResultFileDataSource(cmPlotRequestResponseModel PlotRequestResModel, sToken string, indexValue int64) datamodel.ResourceDataSource {

	var filepath = cmPlotRequestResponseModel.ResultFileInformationModel.FilePath
	var isSeriesFile = cmPlotRequestResponseModel.ResultFileInformationModel.SeriesFile
	var servername = cmPlotRequestResponseModel.ResultFileInformationModel.ServerName

	var pasServerJobModel datamodel.PASServerJobModel
	pasServerJobModel.JobId = cmPlotRequestResponseModel.ResultFileInformationModel.JobId
	pasServerJobModel.JobState = cmPlotRequestResponseModel.ResultFileInformationModel.JobState
	pasServerJobModel.ServerName = cmPlotRequestResponseModel.ResultFileInformationModel.ServerName
	pasServerJobModel.PasURL = cmPlotRequestResponseModel.ResultFileInformationModel.PasUrl

	var id = "res" + strconv.FormatInt(indexValue, 10)
	var ResultDataSource = common.BuildResultDataSource(sToken, id, filepath, isSeriesFile, servername, pasServerJobModel)

	return ResultDataSource

}
func attachDatasourcesToQueries(PlotQueries queries, ResultDataSource datamodel.ResourceDataSource) queries {

	PlotQueries.ResultDataSource = nil
	PlotQueries.ResultDataSource = append(PlotQueries.ResultDataSource, ResultDataSource)
	for i := 0; i < len(PlotQueries.Query); i++ {
		PlotQueries.Query[i].ResultDataSourceRef = nil
		var ResultDataSourceRef resultDataSourceRef
		ResultDataSourceRef.Id = ResultDataSource.Id
		PlotQueries.Query[i].ResultDataSourceRef = append(PlotQueries.Query[i].ResultDataSourceRef, ResultDataSourceRef)

	}

	return PlotQueries

}

func mergeCurveNamesAndPoints(originalPLTModel PlotTemporaryModel,
	overlaidPLTModel PlotTemporaryModel) PlotTemporaryModel {

	var lstMergedCurveNames []string
	var lstLegendNames []string
	var lstMergedCurvePoints [][]float64

	/*
	 * original plt has more points on X Axis choose original plt for X
	 * Axis
	 */
	if len(originalPLTModel.lstCurvesData[0]) >= len(overlaidPLTModel.lstCurvesData[0]) {
		lstMergedCurvePoints = append(lstMergedCurvePoints, originalPLTModel.lstCurvesData[0])
	} else {
		lstMergedCurvePoints = append(lstMergedCurvePoints, overlaidPLTModel.lstCurvesData[0])
	}
	for i := 1; i < len(originalPLTModel.lstCurvesData); i++ {
		lstMergedCurvePoints = append(lstMergedCurvePoints, originalPLTModel.lstCurvesData[i])
	}
	for i := 1; i < len(overlaidPLTModel.lstCurvesData); i++ {
		lstMergedCurvePoints = append(lstMergedCurvePoints, overlaidPLTModel.lstCurvesData[i])
	}

	lstMergedCurveNames = append(lstMergedCurveNames, originalPLTModel.lstCurveNames...)
	lstMergedCurveNames = append(lstMergedCurveNames, overlaidPLTModel.lstCurveNames...)
	lstLegendNames = append(lstLegendNames, originalPLTModel.lstLegendNames...)
	lstLegendNames = append(lstLegendNames, overlaidPLTModel.lstLegendNames...)

	return PlotTemporaryModel{
		lstCurveNames:  lstMergedCurveNames,
		lstCurvesData:  lstMergedCurvePoints,
		lstLegendNames: lstLegendNames,
	}

}

func getRVPPlot(plotQueries queries, ResultFileInformationModel resultFileInformationModel, username string, password string) Res {
	var tempQueries queries
	for i := 0; i < len(plotQueries.ResultDataSource); i++ {
		tempQueries.ResultDataSource = append(tempQueries.ResultDataSource, plotQueries.ResultDataSource[i])
	}
	tempQueries.Query = append(tempQueries.Query, plotQueries.Query[0])

	var lstRVPQueries []Query
	if len(tempQueries.Query[0].RvpPlotDataQuery.RvpPlotColumnInfo.ColumnNames) == 0 {
		for i := 0; i < len(plotQueries.Query); i++ {
			if isRVPQuery(plotQueries.Query[i]) {
				lstRVPQueries = append(lstRVPQueries, plotQueries.Query[i])
				tempQueries.Query[0].RvpPlotDataQuery.RvpPlotColumnInfo.ColumnNames =
					append(tempQueries.Query[0].RvpPlotDataQuery.RvpPlotColumnInfo.ColumnNames, plotQueries.Query[i].RvpPlotDataQuery.RvpPlotColumnInfo.ColumnName)
			}
		}
	} else {
		for i := 0; i < len(plotQueries.Query); i++ {
			if isRVPQuery(plotQueries.Query[i]) {
				lstRVPQueries = append(lstRVPQueries, plotQueries.Query[i])
			}
		}
	}
	return GetRVPPlotData(tempQueries, ResultFileInformationModel, username, password)
}

func isRVPQuery(QueryCtType Query) bool {
	return QueryCtType.RvpPlotDataQuery.RvpPlotColumnInfo.ColumnName != ""
}

func readRVPPLTModel(resultFileName string, pltModel plotRequestResponseModel) PlotTemporaryModel {
	var plotMetaData = pltModel.PlotMetaData

	var resultFileId = pltModel.Queries.Query[0].ResultDataSourceRef[0].Id

	// get plot curve names
	var lstQueries = pltModel.Queries.Query
	var lstLegendNames []string
	var lstCurveNames []string

	var curveName string
	//var strcQuery strcQuery
	var plotQuery Query
	var sLegendName string
	var inlineQueryCTye inlineQuery

	for index := 0; index < len(lstQueries); index++ {
		plotQuery = lstQueries[index]
		/*
		 * Dont add the first query as it is for X Axis
		 */
		if index != 0 {

			if plotQuery.RvpPlotDataQuery.RvpPlotColumnInfo.ColumnName != "" {
				var rvpPlotColumnInfoCType = plotQuery.RvpPlotDataQuery.RvpPlotColumnInfo
				curveName = resultFileName + " : " + rvpPlotColumnInfoCType.ColumnName +
					"(" + plotQuery.VarName + ")" + " :" + resultFileId
				sLegendName = rvpPlotColumnInfoCType.ColumnName + "(" + plotQuery.VarName + ")"
			} else {
				inlineQueryCTye = plotQuery.PlotResultQuery.DataQuery.InlineQuery
				if inlineQueryCTye.Title != "" {
					curveName = resultFileName + ":" + inlineQueryCTye.Title + ":" +
						inlineQueryCTye.Expression + "(" +
						plotQuery.VarName + ")" + " :" + resultFileId
					sLegendName = inlineQueryCTye.Title + "(" + plotQuery.VarName + ")"
				}
			}
			lstCurveNames = append(lstCurveNames, curveName)
			lstLegendNames = append(lstLegendNames, sLegendName)
		}
	}

	var lstListCurvePoints [][]float64

	var lstResponse = pltModel.Responses.Responselist

	for i := 0; i < len(lstResponse); i++ {
		var responseData = lstResponse[i].ResponseData
		var ds = responseData.DataSource
		lstListCurvePoints = append(lstListCurvePoints, ds.Items)
	}

	return PlotTemporaryModel{
		PlotMetaData:   plotMetaData,
		lstCurveNames:  lstCurveNames,
		lstCurvesData:  lstListCurvePoints,
		lstLegendNames: lstLegendNames,
	}

}
