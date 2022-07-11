package template

import (
	"altair/rvs/common"
	"altair/rvs/database"
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	l "altair/rvs/globlog"
	"altair/rvs/toc"
	"altair/rvs/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode"
)

type TemplateTOCModel struct {
	CmPlotRequestResponseModelLst []datamodel.PlotRequestResModel `json:"cmPlotRequestResponseModelLst"`
}

func SaveTemplate(requestdata []byte, username string, token string) (string, error) {

	var plotRequestResModel datamodel.PlotRequestResModel
	json.Unmarshal(requestdata, &plotRequestResModel)

	var lstPlotModel []datamodel.PlotRequestResponseModel

	var plotRequestResponseModelObj = plotRequestResModel.PlotRequestResponseModel
	lstPlotModel = append(lstPlotModel, plotRequestResponseModelObj)

	for i := 0; i < len(lstPlotModel); i++ {
		var plotRequestResponseModelData = lstPlotModel[i]
		var indexValue = utils.GetUniqueRandomIntValue()
		var plotQueries = buildPlotQueries(plotRequestResModel, token, indexValue)
		plotRequestResponseModelData.Queries = plotQueries
	}

	l.Log().Info("Template content is written in temp file, Lets store it in PAS Server")
	templatedatastr, err := json.MarshalIndent(lstPlotModel[0], "", "    ")
	if err != nil {
		return "", err
	}

	var templateDataModel = plotRequestResponseModelObj.TemplateMetaDataModel
	var templateDataObj database.Template
	templateDataObj.ApplicationName = (templateDataModel.ApplicationName)
	templateDataObj.FileExt = templateDataModel.FileExtension
	templateDataObj.FileName = templateDataModel.FileName
	templateDataObj.TemplateData = templatedatastr
	templateDataObj.TemplateName = templateDataModel.TemplateName
	templateDataObj.UserName = username
	templateDataObj.DefaultTemplate = templateDataModel.IsDefault
	templateDataObj.SeriesFile = templateDataModel.IsSeriesFile
	templateDataObj.FilteredReqTemplate = templateDataModel.IsFilteredReqTemplate
	var templateId = utils.GetUniqueRandomIntValue()
	templateDataObj.TemplateId = templateId

	var databaseerr = database.SaveTemplate(templateDataObj)
	templateDataModel.TemplateData = string(templatedatastr)
	templateDataModel.TemplateId = strconv.FormatInt(templateId, 10)

	if databaseerr != nil {
		return "", databaseerr
	}

	if xmlstring, err := json.MarshalIndent(plotRequestResModel, "", "    "); err == nil {
		return string(xmlstring), nil
	}
	return "", nil

}

func buildPlotQueries(cmPlotRequestResponseModel datamodel.PlotRequestResModel, sToken string, indexValue int64) datamodel.Queries {
	var ResultDataSource = buildResultFileDataSource(cmPlotRequestResponseModel, sToken, indexValue)
	var updatedqueries = attachDatasourcesToQueries(cmPlotRequestResponseModel.PlotRequestResponseModel.Queries, ResultDataSource)
	cmPlotRequestResponseModel.PlotRequestResponseModel.Queries = updatedqueries
	return updatedqueries
}

func buildResultFileDataSource(cmPlotRequestResponseModel datamodel.PlotRequestResModel, sToken string, indexValue int64) datamodel.ResourceDataSource {

	var filepath = cmPlotRequestResponseModel.ResultFileInformationModel.FilePath
	var isSeriesFile = cmPlotRequestResponseModel.ResultFileInformationModel.SeriesFile
	var servername = cmPlotRequestResponseModel.ResultFileInformationModel.ServerName

	var pasServerJobModel datamodel.PASServerJobModel
	pasServerJobModel.JobId = cmPlotRequestResponseModel.ResultFileInformationModel.JobId
	pasServerJobModel.JobState = cmPlotRequestResponseModel.ResultFileInformationModel.JobState
	pasServerJobModel.ServerName = cmPlotRequestResponseModel.ResultFileInformationModel.ServerName
	pasServerJobModel.PasURL = cmPlotRequestResponseModel.ResultFileInformationModel.PasUrl

	var id = "res" + strconv.FormatInt(indexValue, 10)
	var ResultDataSource = common.BuildResultDataSource(sToken, id, filepath, isSeriesFile, servername, pasServerJobModel)

	return ResultDataSource

}
func attachDatasourcesToQueries(PlotQueries datamodel.Queries, ResultDataSource datamodel.ResourceDataSource) datamodel.Queries {

	PlotQueries.ResultDataSource = nil
	PlotQueries.ResultDataSource = append(PlotQueries.ResultDataSource, ResultDataSource)
	for i := 0; i < len(PlotQueries.Query); i++ {
		PlotQueries.Query[i].ResultDataSourceRef = nil
		var ResultDataSourceRef datamodel.ResultDataSourceRef
		ResultDataSourceRef.Id = ResultDataSource.Id
		PlotQueries.Query[i].ResultDataSourceRef = append(PlotQueries.Query[i].ResultDataSourceRef, ResultDataSourceRef)

	}

	return PlotQueries

}

func GetSelectedTemplateDetails(servername string, templateid string, isfilterReq bool,
	resultfilepath string, seriesfile string, jobid string, jobstate string, pasURL string, token string, tocReq string) (string, error) {

	l.Log().Info("entering GetSelectedTemplateDetails")

	var templateData, gettemplateerr = database.GetTemplateDetails(templateid)
	if gettemplateerr != nil {
		return "", gettemplateerr
	}

	var plotRequestResponseModel datamodel.PlotRequestResponseModel

	json.Unmarshal([]byte(templateData.TemplateData), &plotRequestResponseModel)

	var templatedatamodel datamodel.TemplateMetaDataModel
	templatedatamodel.ApplicationName = templateData.ApplicationName
	templatedatamodel.FileExtension = templateData.FileExt
	templatedatamodel.FileName = templateData.FileName
	templatedatamodel.IsDefault = templateData.DefaultTemplate
	templatedatamodel.IsFilteredReqTemplate = templateData.FilteredReqTemplate
	templatedatamodel.IsSeriesFile = templateData.SeriesFile
	//templatedatamodel.TemplateData = string(templateData.TemplateData)
	templatedatamodel.TemplateId = strconv.FormatInt(templateData.TemplateId, 10)
	templatedatamodel.TemplateName = templateData.TemplateName
	templatedatamodel.UserName = templateData.UserName

	plotRequestResponseModel.TemplateMetaDataModel = templatedatamodel

	var plotRequestResModel datamodel.PlotRequestResModel

	plotRequestResModel.ApplicationName = templateData.ApplicationName
	plotRequestResModel.PlotRequestResponseModel = plotRequestResponseModel

	var isValid = false
	var isPlotTOC = false
	var plot datamodel.Plots

	json.Unmarshal([]byte(tocReq), &plot)

	if plot.Plot != nil {
		isPlotTOC = true
	}
	if isPlotTOC && isfilterReq {
		isValid = validateTemplateDataForFilterReq(plot, plotRequestResponseModel,
			servername, resultfilepath, seriesfile, jobid, jobstate, token, pasURL)
	} else if isPlotTOC {
		isValid = validateTemplateDataForPlotToc(plot, plotRequestResponseModel)
	} else {

		var rvpToc datamodel.TOCForResultCType
		json.Unmarshal([]byte(tocReq), &rvpToc)
		isValid = validateTemplateDataForRVPToc(rvpToc, plotRequestResponseModel)
	}

	if isValid {

		xmlstring, err := json.MarshalIndent(plotRequestResModel, "", "    ")
		if err != nil {
			return "", err
		}
		return string(xmlstring), nil
	} else {

		return "", &exception.RVSError{
			Errordetails: "",
			Errorcode:    "10050",
			Errortype:    "TYPE_NOT_SUITABLE_TEMPLATE",
		}
	}

}

func SetTemplateAsDefaultTemplate(servername string, fileextension string, sTemplateId string,
	seriesfile bool, username string) (bool, error) {

	var isDefault, err = database.SetTemplateAsDefaultValue(sTemplateId, fileextension, seriesfile, username)
	if err != nil {
		return false, err
	}
	return isDefault, nil

}

func DeleteSelectedTemplate(sTemplateId string) (bool, error) {

	var isDefault, err = database.DeleteTemplateData(sTemplateId)
	if err != nil {
		return false, err
	}
	return isDefault, nil

}

func UpdateTemplate(requestdata []byte, username string, token string) (bool, error) {

	var plotRequestResModel datamodel.PlotRequestResModel
	json.Unmarshal(requestdata, &plotRequestResModel)

	var lstPlotModel []datamodel.PlotRequestResponseModel

	var plotRequestResponseModelObj = plotRequestResModel.PlotRequestResponseModel
	lstPlotModel = append(lstPlotModel, plotRequestResponseModelObj)

	for i := 0; i < len(lstPlotModel); i++ {
		var plotRequestResponseModelData = lstPlotModel[i]
		var indexValue = utils.GetUniqueRandomIntValue()
		var plotQueries = buildPlotQueries(plotRequestResModel, token, indexValue)
		plotRequestResponseModelData.Queries = plotQueries
	}

	templatedatastr, err := json.MarshalIndent(lstPlotModel[0], "", "    ")
	if err != nil {
		return false, err
	}

	var templateDataModel = plotRequestResponseModelObj.TemplateMetaDataModel
	var templateDataObj database.Template
	templateDataObj.ApplicationName = (templateDataModel.ApplicationName)
	templateDataObj.FileExt = templateDataModel.FileExtension
	templateDataObj.FileName = templateDataModel.FileName
	templateDataObj.TemplateData = templatedatastr
	templateDataObj.TemplateName = templateDataModel.TemplateName
	templateDataObj.UserName = username
	templateDataObj.DefaultTemplate = templateDataModel.IsDefault
	templateDataObj.SeriesFile = templateDataModel.IsSeriesFile
	templateDataObj.FilteredReqTemplate = templateDataModel.IsFilteredReqTemplate
	templateDataObj.TemplateId, _ = strconv.ParseInt(templateDataModel.TemplateId, 10, 64)

	var isUpdated, updateerr = database.UpdateTemplateData(templateDataObj)

	if updateerr != nil {
		return false, err
	}
	return isUpdated, nil

}

func GetTemplates(servername string, isfilterReq bool,
	resultfilepath string, seriesfile bool, jobid string, jobstate string, pasURL string, token string, username string,
	fileextension string, tocReq string) (string, error) {

	l.Log().Info("entering GetAllTemplates")

	var templateData, gettemplateerr = database.GetAllTemplateData(fileextension, username, seriesfile)
	if gettemplateerr != nil {
		return "", gettemplateerr
	}

	var templateTocModel = buildTemplateTOCModel(tocReq, isfilterReq, servername, resultfilepath, seriesfile, jobid, jobstate, token, pasURL, templateData, username)

	if xmlstring, err := json.MarshalIndent(templateTocModel, "", "    "); err == nil {
		return string(xmlstring), nil
	}
	return "", nil

}

func buildTemplateTOCModel(sTOCOfResult string, bIsFilterReqs bool, serverName string, sResultFilePath string, sIsSeriesFile bool,
	sJobId string, sJobState string, token string, sPasURL string,
	TemplateDataList []database.Template, username string) TemplateTOCModel {

	var cmrequestResponseModelLst []datamodel.PlotRequestResModel
	var templateTocModel TemplateTOCModel
	for i := 0; i < len(TemplateDataList); i++ {
		var plotModel []datamodel.PlotRequestResponseModel
		var plotRequestResponseModel datamodel.PlotRequestResponseModel
		var templateplotRequestResponseModel datamodel.PlotRequestResponseModel
		var isValid = false

		if TemplateDataList[i].DefaultTemplate {

			json.Unmarshal([]byte(TemplateDataList[i].TemplateData), &templateplotRequestResponseModel)
			var isPlotTOC = false
			var plot datamodel.Plots

			json.Unmarshal([]byte(sTOCOfResult), &plot)

			if plot.Plot != nil {
				isPlotTOC = true
			}

			if isPlotTOC && bIsFilterReqs {
				isValid = validateTemplateDataForFilterReq(plot, templateplotRequestResponseModel,
					serverName, sResultFilePath, strconv.FormatBool(sIsSeriesFile), sJobId, sJobState, token, sPasURL)
			} else if isPlotTOC {
				isValid = validateTemplateDataForPlotToc(plot, templateplotRequestResponseModel)
			} else {
				var rvpToc datamodel.TOCForResultCType
				json.Unmarshal([]byte(sTOCOfResult), &rvpToc)
				isValid = validateTemplateDataForRVPToc(rvpToc, templateplotRequestResponseModel)
			}
			if isValid {
				plotRequestResponseModel = templateplotRequestResponseModel
				plotModel = append(plotModel, plotRequestResponseModel)
			}

		}

		var templateMetaDataModel datamodel.TemplateMetaDataModel
		templateMetaDataModel.TemplateId = strconv.FormatInt(TemplateDataList[i].TemplateId, 10)
		templateMetaDataModel.ApplicationName = TemplateDataList[i].ApplicationName
		templateMetaDataModel.FileExtension = TemplateDataList[i].FileExt
		templateMetaDataModel.FileName = TemplateDataList[i].FileName
		templateMetaDataModel.TemplateName = TemplateDataList[i].TemplateName
		templateMetaDataModel.UserName = TemplateDataList[i].UserName
		templateMetaDataModel.IsDefault = TemplateDataList[i].DefaultTemplate
		templateMetaDataModel.IsSeriesFile = TemplateDataList[i].SeriesFile
		templateMetaDataModel.IsFilteredReqTemplate = TemplateDataList[i].FilteredReqTemplate
		plotRequestResponseModel.TemplateMetaDataModel = templateMetaDataModel

		//add the app templates under the same list of CMPlotRequestResponseModel
		var cmrequestResponseModel datamodel.PlotRequestResModel

		for i := 0; i < len(cmrequestResponseModelLst); i++ {

			if cmrequestResponseModelLst[i].ApplicationName == TemplateDataList[i].ApplicationName {
				cmrequestResponseModel = cmrequestResponseModelLst[i]
				cmrequestResponseModel.ApplicationName = TemplateDataList[i].ApplicationName
				cmrequestResponseModelLst[i].LstPlotRequestResponseModel = append(cmrequestResponseModelLst[i].LstPlotRequestResponseModel,
					plotRequestResponseModel)
			}
		}

		if len(cmrequestResponseModel.LstPlotRequestResponseModel) == 0 {

			if len(plotModel) == 0 {
				plotModel = append(plotModel, plotRequestResponseModel)
			}
			cmrequestResponseModel.ApplicationName = TemplateDataList[i].ApplicationName
			cmrequestResponseModel.LstPlotRequestResponseModel = plotModel
			cmrequestResponseModelLst = append(cmrequestResponseModelLst, cmrequestResponseModel)
		}

	}
	templateTocModel.CmPlotRequestResponseModelLst = cmrequestResponseModelLst
	return templateTocModel
}

func DuplicateTemplateData(servername string, templatename string, templateid string,
	isfilterReqbool bool, resultfilepath string, seriesfilebool bool,
	jobid string, jobstate string, pasURL string, token string, requestdata []byte) (string, error) {

	var templateData, gettemplateerr = database.GetTemplateDetails(templateid)

	if gettemplateerr != nil {
		return "", gettemplateerr
	}

	var plotRequestResponseModel datamodel.PlotRequestResponseModel
	json.Unmarshal([]byte(templateData.TemplateData), &plotRequestResponseModel)

	var templatedatamodel datamodel.TemplateMetaDataModel
	templatedatamodel.ApplicationName = templateData.ApplicationName
	templatedatamodel.FileExtension = templateData.FileExt
	templatedatamodel.FileName = templateData.FileName
	templatedatamodel.TemplateName = templatename
	templatedatamodel.UserName = templateData.UserName
	templatedatamodel.IsDefault = templateData.DefaultTemplate
	templatedatamodel.IsSeriesFile = templateData.SeriesFile
	templatedatamodel.IsFilteredReqTemplate = templateData.FilteredReqTemplate
	plotRequestResponseModel.TemplateMetaDataModel = templatedatamodel

	var plotRequestResModel datamodel.PlotRequestResModel
	plotRequestResModel.ApplicationName = templateData.ApplicationName
	plotRequestResModel.PlotRequestResponseModel = plotRequestResponseModel

	var isValid = false
	var isPlotTOC = false
	var plot datamodel.Plots

	json.Unmarshal(requestdata, &plot)

	if plot.Plot != nil {
		fmt.Println("Plot exists")
		isPlotTOC = true
	}
	if isPlotTOC && isfilterReqbool {
		isValid = validateTemplateDataForFilterReq(plot, plotRequestResponseModel,
			servername, resultfilepath, strconv.FormatBool(seriesfilebool), jobid, jobstate, token, pasURL)
	} else if isPlotTOC {
		fmt.Println("validateTemplateDataForPlotToc")
		isValid = validateTemplateDataForPlotToc(plot, plotRequestResponseModel)
	} else {

		var rvpToc datamodel.TOCForResultCType
		json.Unmarshal(requestdata, &rvpToc)
		isValid = validateTemplateDataForRVPToc(rvpToc, plotRequestResponseModel)
	}

	if isValid {
		templateData.TemplateName = templatename
		var templateId = utils.GetUniqueRandomIntValue()
		templateData.TemplateId = templateId
		templateData.DefaultTemplate = false
		var databaseerr = database.SaveTemplate(templateData)

		if databaseerr != nil {
			return "", databaseerr
		} else {
			plotRequestResModel.PlotRequestResponseModel.TemplateMetaDataModel.TemplateId = strconv.FormatInt(templateData.TemplateId, 10)
			plotRequestResModel.PlotRequestResponseModel.TemplateMetaDataModel.TemplateName = templateData.TemplateName
			plotRequestResModel.PlotRequestResponseModel.TemplateMetaDataModel.IsDefault = false

		}

		if xmlstring, err := json.MarshalIndent(plotRequestResModel, "", "    "); err == nil {
			return string(xmlstring), nil
		}

	} else {

		return "", &exception.RVSError{
			Errordetails: "",
			Errorcode:    "10050",
			Errortype:    "TYPE_NOT_SUITABLE_TEMPLATE",
		}
	}

	return "", nil

}

func validateTemplateDataForFilterReq(plot datamodel.Plots, plotRequestResponseModel datamodel.PlotRequestResponseModel,
	serverName string, sResultFilePath string, sIsSeriesFile string,
	sJobId string, sJobState string, token string, sPasURL string) bool {

	var isSuitable = false
	var subcaseArr = plot.Plot.Subcase
	var listOfSubcaseNames []string
	var mapOfSubcaseAndTypes = make(map[string][]datamodel.TOCType)
	var count = 0
	var subcaseIndex = 0
	for i := 0; i < len(subcaseArr); i++ {
		listOfSubcaseNames = append(listOfSubcaseNames, subcaseArr[i].Name)
	}

	var listOfQueryCTType = plotRequestResponseModel.Queries.Query

	for i := 0; i < len(listOfQueryCTType); i++ {
		var subcaseName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.Subcase.Name
		for j := 0; j < len(subcaseArr); j++ {

			if subcaseArr[j].Name == subcaseName {
				subcaseIndex = subcaseArr[j].Index
			}
		}
		if contains(listOfSubcaseNames, subcaseName) {

			var typeArr []datamodel.TOCType
			var types []string
			var resultTypeName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.Type.Name
			var requestName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name
			var componentName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.Component.Name
			var selectReqNumber = ""
			var selectReqLetter = ""
			for _, c := range requestName {
				if unicode.IsDigit(c) {
					selectReqNumber = selectReqNumber + string(c)
				} else {
					selectReqLetter = selectReqLetter + string(c)
				}
			}
			//get the list of the types for the choosen subcase and check if the saved template has that type present.
			if _, ok := mapOfSubcaseAndTypes[subcaseName]; ok {
				typeArr = mapOfSubcaseAndTypes[subcaseName]
			} else {
				typeArr = GetListOfTypesOfSelectedTOCSubcase(subcaseName, subcaseArr)
				mapOfSubcaseAndTypes[subcaseName] = typeArr
			}
			var startReqName = ""
			var noOfReq int64
			var typeName = ""
			var typeIndex = 0
			for k := 0; k < len(typeArr); k++ {

				types = append(types, typeArr[k].Name)
				if typeArr[k].Name == resultTypeName {
					var requestsOverviewEle = typeArr[k].RequestsOverview
					startReqName = requestsOverviewEle.StartReqName
					noOfReq = int64(requestsOverviewEle.NoOfRequests)
					typeName = resultTypeName
					typeIndex = typeArr[k].Index
				}
			}
			if contains(types, resultTypeName) {
				var sTOCRequest datamodel.TOCRequest
				sTOCRequest.PlotFilter.Subcase.Name = subcaseName
				sTOCRequest.PlotFilter.Subcase.Index = subcaseIndex
				sTOCRequest.PlotFilter.Subcase.Id = ""

				sTOCRequest.PlotFilter.Type.Name = typeName
				sTOCRequest.PlotFilter.Type.Index = typeIndex
				sTOCRequest.PlotFilter.Type.Id = nil

				sTOCRequest.PlotFilter.Filter.Id = nil
				sTOCRequest.PlotFilter.Filter.GetNext = true
				sTOCRequest.PlotFilter.Filter.Start = startReqName
				sTOCRequest.PlotFilter.Filter.Count = noOfReq

				sTOCRequest.PostProcessingType = "PLOT"
				sTOCRequest.Custom = ""
				sTOCRequest.IsCachingRequired = "true"
				sTOCRequest.SchemaVersion = ""
				sTOCRequest.ModelDataSource = ""

				var sFilterTOCOfResultObj, _ = toc.GetPlotToc(serverName, sResultFilePath, sIsSeriesFile,
					sTOCRequest, sJobId, sJobState, token, sPasURL,
					subcaseName, typeName)
				var tocOutput datamodel.Plots
				json.Unmarshal([]byte(sFilterTOCOfResultObj), &tocOutput)
				var reqArr = tocOutput.Plot.Subcase[0].Type[0].Request

				for l := 0; l < len(reqArr); l++ {
					var reqElem = reqArr[l]
					if reqElem.NoOfPoints != 0 {
						var noOfPoints = reqElem.NoOfPoints
						var nameStart = reqElem.NameStart
						if nameStart == requestName {
							var components = GetListOfComponentsForSelectedType(typeArr, resultTypeName)
							for m := 0; m < len(components); m++ {
								var compName = components[m]
								if compName.Name == componentName {
									isSuitable = true
									count = count + 1
								}
							}

						} else {

							var number = ""
							var letter = ""
							for _, c := range nameStart {
								if unicode.IsDigit(c) {
									number = number + string(c)
								} else {
									letter = letter + string(c)
								}
							}

							var selectReqNumberInt, _ = strconv.ParseInt(selectReqNumber, 10, 64)
							var numberInt, _ = strconv.ParseInt(number, 10, 64)
							if selectReqNumberInt >= numberInt &&
								selectReqNumberInt <= numberInt+noOfPoints {
								var components = GetListOfComponentsForSelectedType(typeArr, resultTypeName)
								for x := 0; x < len(components); x++ {
									var compName = components[x].Name
									if compName == componentName {
										isSuitable = true
										count = count + 1
									}
								}

							}
						}
					} else {
						var reqName = reqElem.Name
						if reqName == requestName {
							var components = GetListOfComponentsForSelectedType(typeArr, resultTypeName)
							for y := 0; y < len(components); y++ {
								var compName = components[y].Name

								if compName == componentName {
									isSuitable = true
									count = count + 1
								}
							}

						}
					}
				}
			} else {
				break
			}

		} else {
			break
		}
	}

	if count == len(listOfQueryCTType) {
		isSuitable = true
	} else {
		isSuitable = false
	}

	return isSuitable

}

func validateTemplateDataForPlotToc(plot datamodel.Plots, plotRequestResponseModel datamodel.PlotRequestResponseModel) bool {

	var listOfSubcaseNames []string
	var subcaseArr = plot.Plot.Subcase
	var mapOfSubcaseAndTypes = make(map[string][]datamodel.TOCType)
	var mapOfTypeAndRequests = make(map[string][]datamodel.Request)
	var count = 0
	var isSuitable = false

	for i := 0; i < len(subcaseArr); i++ {
		listOfSubcaseNames = append(listOfSubcaseNames, subcaseArr[i].Name)
	}

	var listOfQueryCTType = plotRequestResponseModel.Queries.Query

	for i := 0; i < len(listOfQueryCTType); i++ {
		var subcaseName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.Subcase.Name
		//check if saved template subcase present in the toc list
		if contains(listOfSubcaseNames, subcaseName) {

			var typeArr []datamodel.TOCType
			var types []string
			var resultTypeName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.Type.Name
			var requestName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.DataRequest.Name
			var componentName = listOfQueryCTType[i].PlotResultQuery.DataQuery.StrcQuery.DistantRequest.Component.Name

			var selectReqNumber = ""
			var selectReqLetter = ""
			for _, c := range requestName {
				if unicode.IsDigit(c) {
					selectReqNumber = selectReqNumber + string(c)
				} else {
					selectReqLetter = selectReqLetter + string(c)
				}
			}

			//get the list of the types for the choosen subcase and check if the saved template has that type present.
			if _, ok := mapOfSubcaseAndTypes[subcaseName]; ok {
				typeArr = mapOfSubcaseAndTypes[subcaseName]
			} else {
				typeArr = GetListOfTypesOfSelectedTOCSubcase(subcaseName, subcaseArr)
				mapOfSubcaseAndTypes[subcaseName] = typeArr
			}

			for j := 0; j < len(typeArr); j++ {
				types = append(types, typeArr[j].Name)
			}
			if contains(types, resultTypeName) {
				var reqArr []datamodel.Request
				//get the list of the types for the choosen subcase and check if the saved template has that type present.
				if _, ok := mapOfTypeAndRequests[resultTypeName]; ok {
					reqArr = mapOfTypeAndRequests[resultTypeName]
				} else {
					reqArr = GetListOfRequestsForSelectedType(resultTypeName, typeArr)
					mapOfTypeAndRequests[resultTypeName] = reqArr
				}
				for k := 0; k < len(reqArr); k++ {

					var reqElem = reqArr[k]
					if reqElem.NoOfPoints != 0 {
						var noOfPoints = reqElem.NoOfPoints
						var nameStart = reqElem.NameStart
						if nameStart == requestName {
							var components = GetListOfComponentsForSelectedType(typeArr, resultTypeName)
							for l := 0; l < len(components); l++ {
								var compName = components[l].Name
								if compName == componentName {
									isSuitable = true
									count = count + 1
								}
							}

						} else {

							var number = ""
							var letter = ""
							for _, c := range nameStart {
								if unicode.IsDigit(c) {
									number = number + string(c)
								} else {
									letter = letter + string(c)
								}
							}
							var selectReqNumberInt, _ = strconv.ParseInt(selectReqNumber, 10, 64)
							var numberInt, _ = strconv.ParseInt(number, 10, 64)
							if selectReqNumberInt >= numberInt &&
								selectReqNumberInt <= numberInt+noOfPoints {
								var components = GetListOfComponentsForSelectedType(typeArr, resultTypeName)
								for x := 0; x < len(components); x++ {
									var compName = components[x].Name
									if compName == componentName {
										isSuitable = true
										count = count + 1
									}
								}

							}
						}
					} else {
						var reqName = reqElem.Name
						if reqName == requestName {
							var components = GetListOfComponentsForSelectedType(typeArr, resultTypeName)
							for y := 0; y < len(components); y++ {
								var compName = components[y].Name

								if compName == componentName {
									isSuitable = true
									count = count + 1
								}
							}

						}
					}
				}
			} else {
				break
			}

		} else {
			break
		}
	}
	if count == len(listOfQueryCTType) {
		isSuitable = true
	} else {
		isSuitable = false
	}

	return isSuitable

}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func GetListOfTypesOfSelectedTOCSubcase(subcaseName string, subcaseArr []datamodel.Subcase) []datamodel.TOCType {
	var typeArr []datamodel.TOCType
	for i := 0; i < len(subcaseArr); i++ {
		var subcaseEle = subcaseArr[i]
		var subName = subcaseEle.Name
		if subName == subcaseName {
			typeArr = subcaseEle.Type
			break
		}
	}

	return typeArr
}

func GetListOfRequestsForSelectedType(resultTypeName string, typeArr []datamodel.TOCType) []datamodel.Request {
	var requestArr []datamodel.Request
	for i := 0; i < len(typeArr); i++ {
		var typeName = typeArr[i].Name
		if typeName == resultTypeName {
			requestArr = typeArr[i].Request
			break
		}
	}

	return requestArr
}

func GetListOfComponentsForSelectedType(typeArr []datamodel.TOCType, resultTypeName string) []datamodel.Component {
	var componentArr []datamodel.Component
	for i := 0; i < len(typeArr); i++ {

		var typeEle = typeArr[i]
		var typeName = typeEle.Name
		if typeName == resultTypeName {
			componentArr = typeEle.Component
			break
		}
	}
	return componentArr
}

func validateTemplateDataForRVPToc(sTOCOfResult datamodel.TOCForResultCType, plotRequestResponseModel datamodel.PlotRequestResponseModel) bool {
	var isSuitable = false
	var columnNamesArr = sTOCOfResult.RVPToc.RVPPlots[0].RvpPlotColumnInfo.ColumnNames

	//get list of columns from the toc result
	var listOfColumns []string
	var count = 0
	for i := 0; i < len(columnNamesArr); i++ {
		listOfColumns = append(listOfColumns, columnNamesArr[i])
	}

	var listOfQueryCTType = plotRequestResponseModel.Queries.Query

	for i := 0; i < len(listOfQueryCTType); i++ {
		var columnName = listOfQueryCTType[i].RvpPlotDataQuery.RvpPlotColumnInfo.ColumnName
		//check if saved template columnName present in the column list
		contains(listOfColumns, columnName)
		count = count + 1
	}

	if count == len(listOfQueryCTType) {
		isSuitable = true
	} else {
		isSuitable = false
	}

	return isSuitable

}
