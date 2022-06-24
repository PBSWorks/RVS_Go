package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const MODEL_COMP_CFG_PATH = "/resources/scripts/GetModelComps.cfg"

func GetModelToc(sModelFilePath string, sJobId string, sJobState string, server string, pasURL string, token string,
	username string, password string) (string, error) {
	fmt.Println("Hello Model!")
	var resulrdatasourceerr error
	var datasource = buildModelDataSource(sModelFilePath, sJobId, sJobState, server, pasURL, token)

	if sJobId == "" && sJobState == "" {
		sModelFilePath, resulrdatasourceerr = common.ResolveFilePortDataSource(datasource, username, password)
		if resulrdatasourceerr != nil {
			return "", resulrdatasourceerr
		}
	} else {
		sModelFilePath, resulrdatasourceerr = common.ResolvePBSPortDataSource(datasource, username, password)
		if resulrdatasourceerr != nil {
			return "", resulrdatasourceerr
		}
	}

	sModelFilePath = strings.Replace(sModelFilePath, common.BACK_SLASH, common.FORWARD_SLASH, -1)

	fileModelFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+common.RM_TOC_XML_FILES, "MODEL")
	fileModelComponents := common.AllocateFile(common.MODEL_COMPONENTS_FILE_NAME, fileModelFolder, username, password)

	extractModelTOC(fileModelComponents, sModelFilePath, username, password)

	b, err := ioutil.ReadFile(fileModelComponents) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	output := string(b)

	result := strings.ReplaceAll(output, "\"Model\": {", "")
	//remove Model close } brace
	result = result[:len(result)-1]

	res, err := common.PrettyString(result)
	if err != nil {
		log.Fatal(err)
	}

	return res, nil

}

func buildModelDataSource(sModelFilePath string, sJobId string, sJobState string, sServerName string,
	pasURL string, token string) datamodel.ResourceDataSource {

	var pasServerJobModel datamodel.PASServerJobModel
	pasServerJobModel.JobId = sJobId
	pasServerJobModel.JobState = sJobState
	pasServerJobModel.ServerName = sServerName
	pasServerJobModel.PasURL = pasURL

	return buildModelFileDataSource(token, sModelFilePath, sServerName, pasServerJobModel)

}

func buildModelFileDataSource(sToken string, sModelFilePath string, servername string,
	pasServerJobModel datamodel.PASServerJobModel) datamodel.ResourceDataSource {

	return common.BuildResultDataSource(sToken, "ds1", sModelFilePath, false, servername,
		pasServerJobModel)
}

func extractModelTOC(sOutputFile string, sModelFilePath string, username string, password string) {

	sCfgFile := common.GetRSHome() + MODEL_COMP_CFG_PATH

	tempCfgFolder := common.AllocateUniqueFolder(common.SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+common.RM_SCRIPT_FILES, "MODEL")
	tempCfgFile := common.AllocateFile(common.GetFileNameWithoutExtension(sOutputFile)+".cfg", tempCfgFolder, username, password)

	sModelCompsFilePath := common.GetRSHome() + common.MODEL_COMP_SOURCE_PATH
	sTempModelFilePath := strings.Replace(sOutputFile, "\\", "/", -1)
	sModelCompsFilePath = strings.Replace(sModelCompsFilePath, "\\", "/", -1)

	readAndWriteModelToFile(sCfgFile, tempCfgFile, sTempModelFilePath, sModelCompsFilePath)
	executeAnimationApplication(tempCfgFile, sModelFilePath, username, password)
	getModelDataForModelFile(sTempModelFilePath, sOutputFile)
}

func readAndWriteModelToFile(sCfgFile string, tempCfgFile string, sTempModelFilePath string, sModelCompsFilePath string) {
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

		if strings.Contains(sLine, TEMP_MODEL_FILE_TAG) {
			iTagIndex := strings.Index(sLine, TEMP_MODEL_FILE_TAG)
			fmt.Fprintf(tempfile, sLine[0:iTagIndex])
			fmt.Fprintf(tempfile, sTempModelFilePath)
			fmt.Fprintf(tempfile, string(sLine[iTagIndex+len(TEMP_MODEL_FILE_TAG)])+"\n")

		} else if strings.Contains(sLine, MODEL_COMPS_SCRIPT_PATH_TAG) {
			iTagIndex := strings.Index(sLine, MODEL_COMPS_SCRIPT_PATH_TAG)
			fmt.Fprintf(tempfile, sLine[0:iTagIndex])
			fmt.Fprintf(tempfile, sModelCompsFilePath)
			fmt.Fprintf(tempfile, string(sLine[iTagIndex+len(MODEL_COMPS_SCRIPT_PATH_TAG)])+"\n")
		} else {
			fmt.Fprintf(tempfile, sLine+"\n")
		}
	}

	originalFile.Close()
	tempfile.Close()

}

func getModelDataForModelFile(modelfilepath string, fileModelComponents string) {
	readModelfile(modelfilepath, fileModelComponents)
}
