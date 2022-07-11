package graph

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	l "altair/rvs/globlog"
	"encoding/json"
	"io/ioutil"
	"os"
)

func SaveInstance(sRequestData []byte, pasURL string, sToken string) (string, error) {
	var instanceSaveModel datamodel.InstanceSaveModel
	var instances datamodel.Plotinstance
	json.Unmarshal(sRequestData, &instanceSaveModel)

	//var filepath = "file=" + instanceSaveModel.ResultFileInformationModel.FilePath
	instanceSaveModel.ResultFileInformationModel.PasUrl = pasURL

	var fileexist = common.DoesFileExist(pasURL, instanceSaveModel.ResultFileInformationModel.JobState,
		instanceSaveModel.ResultFileInformationModel.JobId, sToken, instanceSaveModel.ResultFileInformationModel.FilePath)

	if !instanceSaveModel.CanOverwrite && fileexist {
		return "", &exception.RVSError{
			Errordetails: "File already exist and client did not ask to overwrite the file",
			Errorcode:    "10024",
			Errortype:    "TYPE_OUTPUT_FILE_ALREADY_EXISTS",
		}
	}

	common.CreateFolderIfNotExist(instanceSaveModel.ResultFileInformationModel.ServerName, pasURL, instanceSaveModel.ResultFileInformationModel.JobState,
		instanceSaveModel.ResultFileInformationModel.JobId, sToken, instanceSaveModel.ResultFileInformationModel.FilePath)

	jsonFile, err := os.Open(instanceSaveModel.PlotSaveModelList[0].PlotResponseModel.TemporaryPltFilePath)

	if err != nil {
		l.Log().Error(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &instances)

	instances.Instances.PLT[0].PlotMetaData.UserPreferece.UserPrefereces = append(instances.Instances.PLT[0].PlotMetaData.UserPreferece.UserPrefereces,
		datamodel.UserPrefereces{
			Name:               "user_preferences",
			CurveDatapointSize: USER_DATA_POINT_SIZE,
			CurveLineThickness: USER_LINE_THICKNESS,
			EnableDataPoint:    ENABLE_DATA_POINT,
			CurveColors:        CURVE_DEFAULT_COLOR,
		})

	common.UploadFileWLM(instanceSaveModel.PlotSaveModelList[0].PlotResponseModel.TemporaryPltFilePath, sToken, instanceSaveModel.ResultFileInformationModel.FilePath,
		instanceSaveModel.ResultFileInformationModel.PasUrl, instanceSaveModel.ResultFileInformationModel.JobState,
		instanceSaveModel.ResultFileInformationModel.JobId, instanceSaveModel.CanOverwrite)

	return "", nil

}
