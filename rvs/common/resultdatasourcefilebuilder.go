package common

import (
	"altair/rvs/datamodel"
	"strings"
)

func BuildResultDataSourceBuilder(sToken string, id string, filepath string, isSeriesFile bool, servername string,
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

	resultdatasource.FilePortServer = datamodel.FilePortServer{
		Name:               wlmDetails.ServerName,
		UserName:           wlmDetails.ServerUsername,
		UserPassword:       wlmDetails.Serverpasswd,
		IsSecure:           false,
		Port:               wlmDetails.Serverport,
		AuthorizationToken: sToken,
		PasUrl:             pasServerJobModel.PasURL,
	}

	var lastModifiedTime = GetLastModificationTime(pasServerJobModel.JobState, pasServerJobModel.JobId, pasServerJobModel.PasURL,
		filepath, sToken)

	resultdatasource.LastModificationTime = lastModifiedTime

	return resultdatasource
}
