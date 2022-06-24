package common

import (
	"altair/rvs/datamodel"
	"strings"
)

func BuildPBSDataSourceBuilder(sToken string, id string, filepath string, isSeriesFile bool, servername string,
	pasServerJobModel datamodel.PASServerJobModel) datamodel.ResourceDataSource {

	var resultdatasource datamodel.ResourceDataSource

	resultdatasource.Id = id
	resultdatasource.FilePath = filepath
	resultdatasource.FileIdentifier = filepath
	resultdatasource.SeriesFile = isSeriesFile
	var wlmDetails WLMDetail
	if len(WlmdetailsMap) == 0 {
		var access_token = "access_token=" + strings.Replace(sToken, "Bearer", "", -1)
		GetWLMDetails(access_token, servername, pasServerJobModel.PasURL)
		wlmDetails = WlmdetailsMap[servername]

	} else {
		wlmDetails = WlmdetailsMap[servername]
	}

	resultdatasource.PbsServer = datamodel.PbsServer{
		PasURL:             pasServerJobModel.PasURL,
		Server:             wlmDetails.ServerName,
		Port:               wlmDetails.Serverport,
		JobId:              pasServerJobModel.JobId,
		JobStatus:          pasServerJobModel.JobState,
		UserName:           wlmDetails.ServerUsername,
		UserPassword:       wlmDetails.Serverpasswd,
		IsSecure:           false,
		AuthorizationToken: sToken,
	}

	var lastModifiedTime = GetLastModificationTime(pasServerJobModel.JobState, pasServerJobModel.JobId, pasServerJobModel.PasURL,
		filepath, sToken)

	resultdatasource.LastModificationTime = lastModifiedTime

	return resultdatasource

}
