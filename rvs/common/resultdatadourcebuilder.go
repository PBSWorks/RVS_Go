package common

import (
	"altair/rvs/datamodel"
	"fmt"
)

func BuildResultDataSource(sToken string, index int64, filepath string, isSeriesFile bool, servername string,
	pasServerJobModel datamodel.PASServerJobModel) datamodel.ResourceDataSource {

	if pasServerJobModel.JobId != "" && pasServerJobModel.JobState != "" {
		return BuildPBSDataSourceBuilder(sToken, index, filepath, isSeriesFile, servername,
			pasServerJobModel)
	} else {
		fmt.Println("pas sJobId ", pasServerJobModel.JobId)
		fmt.Println("pas sJobState ", pasServerJobModel.JobState)
		return BuildResultDataSourceBuilder(sToken, index, filepath, isSeriesFile, servername,
			pasServerJobModel)
	}

}
