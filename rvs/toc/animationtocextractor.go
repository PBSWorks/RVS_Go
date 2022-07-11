package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	l "altair/rvs/globlog"
	"altair/rvs/utils"
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func GetAnimationToc(sServerName string, sResultFilePath string, sIsSeriesFile string,
	sTOCRequest datamodel.TOCRequest, sJobId string, sJobState string, token string, pasURL string) (string, error) {

	var username string
	var password string
	var resulrdatasourceerr error

	var datasource = buildAnimTOCRequestForResult(sServerName, sResultFilePath,
		sIsSeriesFile, sJobId, sJobState, token, pasURL)

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

	fileAnimFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+utils.RM_TOC_XML_FILES, "ANIM")
	fileAnimTOCOutput := common.AllocateFile(utils.ANIM_TOC_OUTPUT_FILE_NAME, fileAnimFolder, username, password)
	fileModelFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+utils.RM_TOC_XML_FILES, "ANIM")
	fileModelComponents := common.AllocateFile(utils.MODEL_COMPONENTS_FILE_NAME, fileModelFolder, username, password)

	extractAnimTOC(fileAnimTOCOutput, sResultFilePath, fileModelComponents, username, password)

	aimData, err := ioutil.ReadFile(fileAnimTOCOutput) // just pass the file name
	if err != nil {
		l.Log().Error(err)
	}
	modelData, err := ioutil.ReadFile(fileModelComponents) // just pass the file name
	if err != nil {
		l.Log().Error(err)
	}

	animoutput := string(aimData)
	modeloutput := string(modelData)
	modeloutput = modeloutput[1 : len(modeloutput)-1]
	var json_string = "\"Plot\" : null , \"rvpToc\" : null, \"SupportedPPType\" : \"ALL\",\"Custom\" : null"

	var output = "{" + animoutput + "," + modeloutput + "," + json_string + "}"

	res, err := utils.PrettyString(output)
	if err != nil {
		log.Fatal(err)
	}
	return res, nil

}

func buildAnimTOCRequestForResult(sServerName string, sResultFilePath string, sIsSeriesFile string,
	sJobId string, sJobState string, token string, pasURL string) datamodel.ResourceDataSource {

	var pasServerJobModel datamodel.PASServerJobModel
	pasServerJobModel.JobId = sJobId
	pasServerJobModel.JobState = sJobState
	pasServerJobModel.ServerName = sServerName
	pasServerJobModel.PasURL = pasURL

	var index = utils.GetUniqueRandomIntValue()
	var isSeriesFile, _ = strconv.ParseBool(sIsSeriesFile)
	return buildAnimResultFileDataSource(token, index, sResultFilePath, isSeriesFile, sServerName, pasServerJobModel)

}

func buildAnimResultFileDataSource(sToken string, index int64, filepath string, isSeriesFile bool, servername string,
	pasServerJobModel datamodel.PASServerJobModel) datamodel.ResourceDataSource {
	var id = "res" + strconv.FormatInt(index, 10)
	return common.BuildResultDataSource(sToken, id, filepath, isSeriesFile, servername, pasServerJobModel)
}

func extractAnimTOC(sOutputFile string, sResultFilePath string, fileModelComponents string, username string, password string) {

	sCfgFile := utils.GetRSHome() + utils.SCRIPTS + utils.ANIM_TOC_CFG_PATH

	tempCfgFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+utils.RM_SCRIPT_FILES, "ANIM")
	tempCfgFile := common.AllocateFile(utils.GetFileNameWithoutExtension(sOutputFile)+".cfg", tempCfgFolder, username, password)

	sModelCompsFilePath := utils.GetRSHome() + utils.SCRIPTS + utils.MODEL_COMP_SOURCE_FILE_NAME
	sTocOutputFile := strings.Replace(sOutputFile, "\\", "/", -1)
	sTempModelFilePath := sTocOutputFile + utils.MODEL_FILE_EXT
	sModelCompsFilePath = strings.Replace(sModelCompsFilePath, "\\", "/", -1)

	readAndWriteToFile(sCfgFile, tempCfgFile, sTocOutputFile, sModelCompsFilePath, sTempModelFilePath)
	executeAnimationApplication(tempCfgFile, sResultFilePath, username, password)
	getModelData(sTempModelFilePath, fileModelComponents)
}

func readAndWriteToFile(sCfgFile string, tempCfgFile string, sTocOutputFile string, sModelCompsFilePath string, sTempModelFilePath string) {
	originalFile, err := os.Open(sCfgFile)

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	tempfile, err := os.Create(tempCfgFile)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	scanner := bufio.NewScanner(originalFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {

		sLine := scanner.Text()

		if strings.Contains(sLine, utils.ANIM_TOC_MARKER_TAG) {
			iTagIndex := strings.Index(sLine, utils.ANIM_TOC_MARKER_TAG)
			fmt.Fprintf(tempfile, sLine[0:iTagIndex])
			fmt.Fprintf(tempfile, sTocOutputFile)
			fmt.Fprintf(tempfile, string(sLine[iTagIndex+len(utils.ANIM_TOC_MARKER_TAG)])+"\n")

		} else if strings.Contains(sLine, utils.MODEL_COMPS_SCRIPT_PATH_TAG) {
			iTagIndex := strings.Index(sLine, utils.MODEL_COMPS_SCRIPT_PATH_TAG)
			fmt.Fprintf(tempfile, sLine[0:iTagIndex])
			fmt.Fprintf(tempfile, sModelCompsFilePath)
			fmt.Fprintf(tempfile, string(sLine[iTagIndex+len(utils.MODEL_COMPS_SCRIPT_PATH_TAG)])+"\n")
		} else if strings.Contains(sLine, utils.TEMP_MODEL_FILE_TAG) {
			iTagIndex := strings.Index(sLine, utils.TEMP_MODEL_FILE_TAG)
			fmt.Fprintf(tempfile, sLine[0:iTagIndex])
			fmt.Fprintf(tempfile, sTempModelFilePath)
			fmt.Fprintf(tempfile, string(sLine[iTagIndex+len(utils.TEMP_MODEL_FILE_TAG)])+"\n")
		} else {
			fmt.Fprintf(tempfile, sLine+"\n")

		}
	}

	originalFile.Close()
	tempfile.Close()

}

func getModelData(modelfilepath string, fileModelComponents string) {
	readModelfile(modelfilepath, fileModelComponents)
}
