package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	l "altair/rvs/globlog"
	"altair/rvs/utils"
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func GetPlotToc(sServerName string, sResultFilePath string, sIsSeriesFile string,
	sTOCRequest datamodel.TOCRequest, sJobId string, sJobState string, token string, pasURL string,
	sSubcaseName string, sTypeName string) (string, error) {

	var bFetchFilteredTOC = false
	var username string
	var password string
	var resulrdatasourceerr error

	var datasource = buildTOCRequestForResult(sServerName, sResultFilePath,
		sIsSeriesFile, sJobId, sJobState, token, pasURL)

	outputFileFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+utils.RM_TOC_XML_FILES, "PLOT")
	sTOCOutputFile := common.AllocateFile(utils.PLOT_TOC_OUTPUT_FILE_NAME_PART, outputFileFolder, username, password)
	sTOCOutputFile = utils.GetPlatformIndependentFilePath(sTOCOutputFile, false)
	plotOmlFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+utils.RM_SCRIPT_FILES, "PLOT")
	plotOmlFile := common.AllocateFile(utils.PLOT_TOC_OUTPUT_OML_NAME_PART, plotOmlFolder, username, password)

	if utils.IsValidString(sSubcaseName) && utils.IsValidString(sTypeName) {
		l.Log().Info("Request to fetch filter Plot TOC for subcase [%s] and result type [%s]",
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
	sResultFilePath = strings.Replace(sResultFilePath, utils.BACK_SLASH, utils.FORWARD_SLASH, -1)
	writeIntoOmlFile(plotOmlFile, sResultFilePath, sTOCOutputFile, utils.GetRSHome()+utils.SCRIPTS, bFetchFilteredTOC, sSubcaseName, sTypeName)

	ExecuteComposeApplicatopn(plotOmlFile, username, password)

	b, err := ioutil.ReadFile(sTOCOutputFile) // just pass the file name
	if err != nil {
		l.Log().Error(err)
	}
	output := string(b)

	size := len([]rune(output))
	if size > 2097152 {
		output = readTOCAndWriteFilterTOC(sTOCOutputFile)
	}

	res, err := utils.PrettyString(output)
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

	var index = utils.GetUniqueRandomIntValue()
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

	var firstline = "clc; clear; close all; tic();" + utils.NEWLINE +
		"global HWEP_RS_RESULTFILE;" + utils.NEWLINE + "global SEQUENTIAL_REQ_END_INDEX;" +
		utils.NEWLINE + "HWEP_RS_RESULTFILE = " + utils.SINGLE_QUOTE + sResultFilePath + utils.SINGLE_QUOTE

	var secondline = "global HWEP_RS_TOC_OUTPUTFILE;" + utils.NEWLINE + "HWEP_RS_TOC_OUTPUTFILE = " +
		utils.SINGLE_QUOTE + sTOCOutputFile + utils.SINGLE_QUOTE

	var thirdline = "status = addpath (" + utils.SINGLE_QUOTE + sTOCMasterOmlFileName + utils.SINGLE_QUOTE + ");"
	var forthline = "run (" + utils.SINGLE_QUOTE + utils.PLOT_TOC_OML_FILE_NAME + utils.SINGLE_QUOTE + ");"

	var fifthline string
	if bFetchFilteredTOC {
		fifthline = "getFilteredTOC(" + utils.SINGLE_QUOTE + sSubcaseName + utils.SINGLE_QUOTE + "," +
			utils.SINGLE_QUOTE + sTypeName + utils.SINGLE_QUOTE + ")" + utils.NEWLINE
	} else {
		fifthline = "getTOC" + utils.NEWLINE
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
