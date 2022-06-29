package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"encoding/json"
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

func GetRVPToc(fileInformationModel datamodel.FileInformationModel, sToken string, username string, password string) (string, error) {

	var sJobState = fileInformationModel.JobState
	var sJobId = fileInformationModel.JobId
	var sPASURL = fileInformationModel.PasUrl
	var FilePath = fileInformationModel.FilePath
	var sServerName = fileInformationModel.ServerName

	//	var lastModTIme = common.GetLastModificationTime(sJobState, sJobId, sPASURL, FilePath, sToken)

	var datasource = buildRVPTOCRequestForResult(sServerName, FilePath,
		"false", sJobId, sJobState, sToken, sPASURL)

	var tocRequestForResult datamodel.TOCRequest
	tocRequestForResult.IsCachingRequired = "true"
	tocRequestForResult.PostProcessingType = "RVP_PLOT"
	tocRequestForResult.ResultDataSource = datasource
	var sRVPResultFilePath = ""
	var resulrdatasourceerr error
	if sJobId == "" && sJobState == "" {
		sRVPResultFilePath, resulrdatasourceerr = common.ResolveFilePortDataSource(datasource, username, password)
		if resulrdatasourceerr != nil {
			return "", resulrdatasourceerr
		}
	} else {
		sRVPResultFilePath, resulrdatasourceerr = common.ResolvePBSPortDataSource(datasource, username, password)
		if resulrdatasourceerr != nil {
			return "", resulrdatasourceerr
		}
	}

	sRVPResultFilePath = strings.Replace(sRVPResultFilePath, common.BACK_SLASH, common.FORWARD_SLASH, -1)

	fileExtension := filepath.Ext(sRVPResultFilePath)

	var rvpFileModeldata = getRVPFileModel(fileExtension, common.RvpFilesModel)

	var rvpProcessDataModel RVPProcessDataModel
	rvpProcessDataModel.RvpResultFilePath = sRVPResultFilePath
	rvpProcessDataModel.RvpFileModel = rvpFileModeldata
	rvpProcessDataModel.RvpFileExtension = fileExtension

	var rvpPlot = runDataExtractor(rvpProcessDataModel)

	if xmlstring, err := json.MarshalIndent(rvpPlot, "", "    "); err == nil {
		return string(xmlstring), nil
	} else {

		return "", nil
	}
}

func buildRVPTOCRequestForResult(sServerName string, sResultFilePath string, sIsSeriesFile string,
	sJobId string, sJobState string, token string, pasURL string) datamodel.ResourceDataSource {

	var pasServerJobModel datamodel.PASServerJobModel
	pasServerJobModel.JobId = sJobId
	pasServerJobModel.JobState = sJobState
	pasServerJobModel.ServerName = sServerName
	pasServerJobModel.PasURL = pasURL

	var index = common.GetUniqueRandomIntValue()
	var isSeriesFile, _ = strconv.ParseBool(sIsSeriesFile)
	return buildRVPResultFileDataSource(token, index, sResultFilePath, isSeriesFile, sServerName, pasServerJobModel)

}

func buildRVPResultFileDataSource(sToken string, index int64, filepath string, isSeriesFile bool, servername string,
	pasServerJobModel datamodel.PASServerJobModel) datamodel.ResourceDataSource {
	var id = "res" + strconv.FormatInt(index, 10)
	return common.BuildResultDataSource(sToken, id, filepath, isSeriesFile, servername, pasServerJobModel)
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

func runDataExtractor(rvpGetTOCModel RVPProcessDataModel) datamodel.TOCForResultCType {
	var tocForResult datamodel.TOCForResultCType

	if strings.HasSuffix(rvpGetTOCModel.RvpFileExtension, common.RVP_FILE_EXTENSION) {
		tocForResult.RVPToc.RVPPlots = append(tocForResult.RVPToc.RVPPlots, RVPFileTOCExtractor(rvpGetTOCModel.RvpResultFilePath))

	} else {
		tocForResult.RVPToc.RVPPlots = append(tocForResult.RVPToc.RVPPlots, GenericFileTOCExtractor(rvpGetTOCModel))
	}
	return tocForResult

}
