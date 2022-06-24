package common

import (
	"altair/rvs/datamodel"
)

func BuildResultDataSource(sToken string, id string, filepath string, isSeriesFile bool, servername string,
	pasServerJobModel datamodel.PASServerJobModel) datamodel.ResourceDataSource {

	if pasServerJobModel.JobId != "" && pasServerJobModel.JobState != "" {

		return BuildPBSDataSourceBuilder(sToken, id, filepath, isSeriesFile, servername,
			pasServerJobModel)
	} else {
		return BuildResultDataSourceBuilder(sToken, id, filepath, isSeriesFile, servername,
			pasServerJobModel)
	}

}
