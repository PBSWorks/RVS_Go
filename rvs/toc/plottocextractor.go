package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const PLOT_TOC_OUTPUT_FILE_NAME_PART = "PlotTOC.json"
const PLOT_TOC_OUTPUT_OML_NAME_PART = "PlotTOC.oml"
const PLOT_TOC_OML_FILE_NAME = "GetPlotTOC.oml"

func GetPlotToc(sServerName string, sResultFilePath string, sIsSeriesFile string,
	sTOCRequest datamodel.TOCRequest, sJobId string, sJobState string, token string, pasURL string,
	sSubcaseName string, sTypeName string) (string, error) {

	var bFetchFilteredTOC = false
	var username string
	var password string
	var resulrdatasourceerr error

	var datasource = buildTOCRequestForResult(sServerName, sResultFilePath,
		sIsSeriesFile, sJobId, sJobState, token, pasURL)

	outputFileFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+common.RM_TOC_XML_FILES, "PLOT")
	sTOCOutputFile := common.AllocateFile(PLOT_TOC_OUTPUT_FILE_NAME_PART, outputFileFolder, username, password)
	sTOCOutputFile = common.GetPlatformIndependentFilePath(sTOCOutputFile, false)
	plotOmlFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+common.RM_SCRIPT_FILES, "PLOT")
	plotOmlFile := common.AllocateFile(PLOT_TOC_OUTPUT_OML_NAME_PART, plotOmlFolder, username, password)

	if common.IsValidString(sSubcaseName) && common.IsValidString(sTypeName) {
		log.Printf("Request to fetch filter Plot TOC for subcase [%s] and result type [%s]",
			sTOCRequest.PlotFilter.Subcase.Name, sTOCRequest.PlotFilter.Type.Name)
		bFetchFilteredTOC = true
	}
	if sJobId == "" && sJobState == "" {
		sResultFilePath, resulrdatasourceerr = common.ResolveFilePortDataSource(datasource, username, password)
		if resulrdatasourceerr != nil {
			return "", resulrdatasourceerr
		}
	} else {
		sResultFilePath, resulrdatasourceerr = common.ResolvePBSPortDataSource(datasource, username, password)
		if resulrdatasourceerr != nil {
			return "", resulrdatasourceerr
		}
	}
	sResultFilePath = strings.Replace(sResultFilePath, common.BACK_SLASH, common.FORWARD_SLASH, -1)
	writeIntoOmlFile(plotOmlFile, sResultFilePath, sTOCOutputFile, common.GetRSHome()+common.TOC_MASTER_OML_FILE_NAME, bFetchFilteredTOC, sSubcaseName, sTypeName)

	ExecuteComposeApplicatopn(plotOmlFile, username, password)

	b, err := ioutil.ReadFile(sTOCOutputFile) // just pass the file name
	if err != nil {
		log.Print(err)
	}
	output := string(b)

	size := len([]rune(output))
	if size > 2097152 {
		output = readTOCAndWriteFilterTOC(sTOCOutputFile)
	}

	res, err := common.PrettyString(output)
	if err != nil {
		log.Fatal(err)
	}

	return res, nil
}

func buildTOCRequestForResult(sServerName string, sResultFilePath string, sIsSeriesFile string,
	sJobId string, sJobState string, token string, pasURL string) datamodel.ResourceDataSource {

	var pasServerJobModel datamodel.PASServerJobModel
	pasServerJobModel.JobId = sJobId
	pasServerJobModel.JobState = sJobState
	pasServerJobModel.ServerName = sServerName
	pasServerJobModel.PasURL = pasURL

	var index = common.GetUniqueRandomIntValue()
	var isSeriesFile, _ = strconv.ParseBool(sIsSeriesFile)
	return buildResultFileDataSource(token, index, sResultFilePath, isSeriesFile, sServerName, pasServerJobModel)

}

func buildResultFileDataSource(sToken string, index int64, filepath string, isSeriesFile bool, servername string,
	pasServerJobModel datamodel.PASServerJobModel) datamodel.ResourceDataSource {
	var id = "res" + strconv.FormatInt(index, 10)
	return common.BuildResultDataSource(sToken, id, filepath, isSeriesFile, servername, pasServerJobModel)
}

func writeIntoOmlFile(tempOmlFile string, sResultFilePath string,
	sTOCOutputFile string, sTOCMasterOmlFileName string, bFetchFilteredTOC bool, sSubcaseName string, sTypeName string) {

	var firstline = "clc; clear; close all; tic();" + common.NEWLINE +
		"global HWEP_RS_RESULTFILE;" + common.NEWLINE + "global SEQUENTIAL_REQ_END_INDEX;" +
		common.NEWLINE + "HWEP_RS_RESULTFILE = " + common.SINGLE_QUOTE + sResultFilePath + common.SINGLE_QUOTE

	var secondline = "global HWEP_RS_TOC_OUTPUTFILE;" + common.NEWLINE + "HWEP_RS_TOC_OUTPUTFILE = " + common.SINGLE_QUOTE + sTOCOutputFile + common.SINGLE_QUOTE

	var thirdline = "status = addpath (" + common.SINGLE_QUOTE + sTOCMasterOmlFileName + common.SINGLE_QUOTE + ");"
	var forthline = "run (" + common.SINGLE_QUOTE + PLOT_TOC_OML_FILE_NAME + common.SINGLE_QUOTE + ");"

	var fifthline string
	if bFetchFilteredTOC {
		fifthline = "getFilteredTOC(" + common.SINGLE_QUOTE + sSubcaseName + common.SINGLE_QUOTE + "," + common.SINGLE_QUOTE + sTypeName + common.SINGLE_QUOTE + ")" + common.NEWLINE
	} else {
		fifthline = "getTOC" + common.NEWLINE
	}

	plottocfilecontent := []string{firstline, secondline, thirdline, forthline, fifthline}

	file, err := os.OpenFile(tempOmlFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	datawriter := bufio.NewWriter(file)

	for _, data := range plottocfilecontent {
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	file.Close()
}
