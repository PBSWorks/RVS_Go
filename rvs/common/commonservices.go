package common

import (
	"altair/rvs/datamodel"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

var Plotplugin datamodel.Plugin
var Animationplugin datamodel.Plugin
var RVPplugin datamodel.Plugin
var RVPDataplugin datamodel.Plugin
var SupportedFilePatternOutput *datamodel.SupportedFilePatternOutput
var seriespatternFile *datamodel.SeriespatternFile

type WLMDetail struct {
	ServerName     string `json:"serverName"`
	Serverport     string `json:"serverport"`
	ServerUsername string `json:"serverUsername"`
	Serverpasswd   string `json:"serverpasswd"`
	ObjectId       string `json:"objectId"`
	PasURL         string `json:"pasURL"`
}

var servicedata embed

type embed struct {
	Embedded embedded `json:"_embedded"`
}
type embedded struct {
	Service []service `json:"service"`
}

type service struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Username string `json:"username"`
	ObjectId string `json:"objectId"`
}

type createfolderres struct {
	Success  string `json:"success"`
	Data     string `json:"data"`
	StdErr   string `json:"stdErr"`
	ExitCode string `json:"exitCode"`
}

type lastMoifiedoutput struct {
	Success  bool   `json:"success"`
	Data     data   `json:"data"`
	ExitCode string `json:"exitCode"`
}

type data struct {
	Files      []files `json:"files"`
	TotalFiles int     `json:"totalFiles"`
}

type files struct {
	Modified int64  `json:"modified"`
	Filename string `json:"filename"`
	FileExt  string `json:"fileExt"`
}

func GetWLMDetails(cookies string, wlmName string, pasUrl string) {
	if !IsValidString(wlmName) {
		fmt.Println("WLM Name is empty wlmName=" + wlmName)
	}
	wlmName = strings.TrimSpace(wlmName)
	fecthAllWLM(cookies, wlmName, pasUrl)
}

func fecthAllWLM(sCookie string, pasName string, pasUrl string) {
	serviceurl := "/storage/service"
	serviceurl = SiteConfigData.RMServers[0].PAServerURL + serviceurl
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("GET", serviceurl, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Cookie", sCookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal([]byte(body), &servicedata)
	for k := range WlmdetailsMap {
		delete(WlmdetailsMap, k)
	}

	for i := 0; i < len(servicedata.Embedded.Service); i++ {

		u, _ := url.Parse(servicedata.Embedded.Service[0].Url)

		host, port, _ := net.SplitHostPort(u.Host)

		WlmdetailsMap[servicedata.Embedded.Service[i].Name] = WLMDetail{
			ServerName:     host,
			Serverport:     port,
			ServerUsername: servicedata.Embedded.Service[i].Username,
			Serverpasswd:   "pbsadmin",
			ObjectId:       servicedata.Embedded.Service[i].ObjectId,
			PasURL:         pasUrl,
		}
	}

}

func DownloadFileWLM(pasUrl string, jobstate string, jobId string, filepath string, sToken string) string {
	var urlString = buildFileDownloadURL(pasUrl, jobstate, jobId, filepath)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", sToken)

	if jobstate == "R" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("charset", "utf-8")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func buildFileDownloadURL(fileDownloadUrl string, jobstate string, jobId string, filepath string) string {

	if jobstate == "R" && strings.Contains(fileDownloadUrl, PAS_URL_VALUE) {
		fileDownloadUrl = strings.Replace(fileDownloadUrl, PAS_URL_VALUE, JOB_OPERATION, -1)
		fileDownloadUrl = fileDownloadUrl + "/files/download?serversidefilepath=" + filepath
	} else {
		var encodedpath = url.QueryEscape(filepath)
		fileDownloadUrl = fileDownloadUrl + REST_SERVICE_URL + "/files/download?serversidefilepath=" + encodedpath
	}

	if jobId != "" {
		fileDownloadUrl = fileDownloadUrl + "&jobid=" + jobId
	}

	return fileDownloadUrl
}

func DoesFileExist(pasUrl string, jobstate string, jobId string, sToken string, pathToCheck string) bool {

	var fileExist bool
	var sUrl = buildFileCheckURL(pasUrl, jobstate, sToken)
	var postData = buildPostData(pathToCheck, jobId)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", sUrl, bytes.NewBufferString(postData))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", sToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if string(body) != "" {
		fileExist = parseFileExists(body, pathToCheck)
	}

	return fileExist
}

func buildFileCheckURL(pasUrl string, jobstate string, sToken string) string {

	if jobstate == "R" && strings.Contains(pasUrl, PAS_URL_VALUE) {
		pasUrl = strings.Replace(pasUrl, PAS_URL_VALUE, JOB_OPERATION, -1)
	} else {
		pasUrl = pasUrl + REST_SERVICE_URL
	}

	pasUrl = pasUrl + "/files/file/exists"
	return pasUrl
}

func buildPostData(pathToCheck string, jobId string) string {

	var json = "{\r\n"
	if jobId != "" {
		json = json + "\"jobid\": \"" + jobId + "\",\r\n"
	}

	json = json + " \"paths\": [\"" + pathToCheck + "\"]\r\n" + "}"
	return json
}

func parseFileExists(returnedData []byte, pathtocheck string) bool {

	var payload interface{}
	json.Unmarshal(returnedData, &payload)
	m := payload.(map[string]interface{})

	iter := reflect.ValueOf(m["data"]).MapRange()
	for iter.Next() {
		value := iter.Value().Interface()
		valueiter := reflect.ValueOf(value).MapRange()
		for valueiter.Next() {
			key := valueiter.Key().Interface()
			if key == "fileExists" {
				fileexistvalue := fmt.Sprint(valueiter.Value().Interface())
				fileexist, _ := strconv.ParseBool(fileexistvalue)
				return fileexist
			}
		}
	}

	return false
}

func CreateFolderIfNotExist(servername string, pasUrl string, jobstate string, jobId string, sToken string, filepath string) error {
	//var createdDirFilePath string
	var jobLocation = path.Dir(filepath)
	var folderExist = DoesFileExist(pasUrl, jobstate, jobId, sToken, jobLocation)
	if !folderExist {
		var exitcode = createDirWLM(servername, pasUrl, jobstate, jobId, filepath, sToken)
		if !(exitcode == 0) {
			var msg = "Directory does not exists, Choose another directory"
			return errors.New(msg)
		}
	}
	return errors.New("")

}

func createDirWLM(servername string, pasUrl string, jobstate string, jobId string, filepath string, sToken string) int {
	var sCreateDirUrl = BuildCreateDirURL(servername, pasUrl, jobstate, sToken)
	var postData = buildPostData(filepath, jobId)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", sCreateDirUrl, bytes.NewBufferString(postData))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", sToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var exitCode = parseJsonCreateDir(body)
	return exitCode

}

func BuildCreateDirURL(servername string, pasUrl string, jobstate string, sToken string) string {

	GetWLMDetails(sToken, servername, pasUrl)

	var sPASURL = "https://" + WlmdetailsMap[servername].ServerName + ":" + WlmdetailsMap[servername].Serverport

	if jobstate == "R" {
		sPASURL = sPASURL + JOB_OPERATION
	} else {
		sPASURL = sPASURL + PAS_URL_VALUE + REST_SERVICE_URL
	}
	sPASURL = sPASURL + "/files/dir/create"
	return sPASURL
}

func parseJsonCreateDir(response []byte) int {

	var Createfolderres createfolderres
	json.Unmarshal(response, &Createfolderres)

	i, _ := strconv.Atoi(Createfolderres.ExitCode)
	return i

}

func UploadFileWLM(filename string, sToken string, filepath string, pasUrl string, jobstate string, jobId string, canOverWrite bool) error {

	//var boundry = "****" + "DABSADSAJN" + "****"
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("attfile", filepath)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	contentType := bodyWriter.FormDataContentType()

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	bodyWriter.Close()
	var targetUrl = buildFileUploadUrl(sToken, filepath, pasUrl, jobstate, jobId, canOverWrite)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", targetUrl, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", sToken)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))
	return nil
}

func buildFileUploadUrl(sToken string, filepath string, pasUrl string, jobstate string, jobId string, canOverWrite bool) string {
	var encodedpath = url.QueryEscape(filepath)
	var sPASURL = pasUrl

	if jobstate == "R" && strings.Contains(pasUrl, PAS_URL_VALUE) {
		sPASURL = strings.Replace(pasUrl, PAS_URL_VALUE, JOB_OPERATION, -1)
	} else {
		sPASURL = sPASURL + REST_SERVICE_URL
	}

	var baseUrl = sPASURL + "/files/upload?serversidefilepath=" + encodedpath

	log.Printf("buildFileUploadUrl baseUrl: " + baseUrl)

	if jobId != "" {
		baseUrl = baseUrl + "&jobid=" + jobId
	}
	baseUrl = baseUrl + "&override=" + strconv.FormatBool(canOverWrite)

	return baseUrl
}

func GetSupportedFilePatternsForAllServers(sToken string) string {

	SupportedFilePatternOutput = new(datamodel.SupportedFilePatternOutput)

	var data = make(map[string]map[string]bool)
	data["PLOT"] = getPlotSupportedFiles()
	data["ANIMATION"] = getAnimationSupportedFiles()
	data["RVP_PLOT"] = getRVPSupportedFiles()
	data["DIRECT_PLOT"] = getRVPDataSupportedFiles()

	var lstAnimPatterns []string
	var lstplotPatterns []string
	var lstRVPPlotPatterns []string
	var lstRVPDataPlotPatterns []string

	var lstCommonPatterns []string

	for i := 0; i < len(Animationplugin.DataProvider.SupportedFiles.File); i++ {
		lstAnimPatterns = append(lstAnimPatterns, Animationplugin.DataProvider.SupportedFiles.File[i].Value)
	}
	for j := 0; j < len(Plotplugin.DataProvider.SupportedFiles.File); j++ {
		lstplotPatterns = append(lstplotPatterns, Plotplugin.DataProvider.SupportedFiles.File[j].Value)
	}
	for k := 0; k < len(RVPplugin.DataProvider.SupportedFiles.File); k++ {
		lstRVPPlotPatterns = append(lstRVPPlotPatterns, RVPplugin.DataProvider.SupportedFiles.File[k].Pattern)
	}
	for key, _ := range data["DIRECT_PLOT"] {
		lstRVPDataPlotPatterns = append(lstRVPDataPlotPatterns, key)
	}

	for a := 0; a < len(Animationplugin.DataProvider.SupportedFiles.File); a++ {
		if contains(lstplotPatterns, Animationplugin.DataProvider.SupportedFiles.File[a].Value) {
			lstCommonPatterns = append(lstCommonPatterns, Animationplugin.DataProvider.SupportedFiles.File[a].Value)
		}
	}
	var SupportedFilePatterns = new(datamodel.SupportedFilePatterns)
	for i := 0; i < len(lstCommonPatterns); i++ {
		var postProcessingType string
		if data["ANIMATION"][lstCommonPatterns[i]] {
			postProcessingType = "ANIMATION"
		} else {
			postProcessingType = "PLOT"
		}

		SupportedFilePatterns.Pattern = append(SupportedFilePatterns.Pattern, datamodel.Pattern{
			Value:                     lstCommonPatterns[i],
			PpType:                    "ALL",
			DefaultPostProcessingType: postProcessingType,
		})

		lstAnimPatterns = remove(lstAnimPatterns, lstCommonPatterns[i])
		lstplotPatterns = remove(lstplotPatterns, lstCommonPatterns[i])
	}

	for _, sPattern := range lstAnimPatterns {
		SupportedFilePatterns.Pattern = append(SupportedFilePatterns.Pattern, datamodel.Pattern{
			Value:                     sPattern,
			PpType:                    "ANIMATION",
			DefaultPostProcessingType: "ANIMATION",
		})
	}

	for _, sPattern := range lstplotPatterns {
		SupportedFilePatterns.Pattern = append(SupportedFilePatterns.Pattern, datamodel.Pattern{
			Value:                     sPattern,
			PpType:                    "PLOT",
			DefaultPostProcessingType: "PLOT",
		})
	}

	for _, sPattern := range lstRVPPlotPatterns {
		SupportedFilePatterns.Pattern = append(SupportedFilePatterns.Pattern, datamodel.Pattern{
			Value:                     sPattern,
			PpType:                    "RVP_PLOT",
			DefaultPostProcessingType: "RVP_PLOT",
		})
	}

	for _, sPattern := range lstRVPDataPlotPatterns {
		SupportedFilePatterns.Pattern = append(SupportedFilePatterns.Pattern, datamodel.Pattern{
			Value:                     sPattern,
			PpType:                    "DIRECT_PLOT",
			DefaultPostProcessingType: "DIRECT_PLOT",
		})
	}

	SupportedFilePatternOutput.MapFilePatterns.WLMSERVER.IsHWConfigured, _ = isHWInstalledandConfigured()
	SupportedFilePatternOutput.MapFilePatterns.WLMSERVER.Pattern = SupportedFilePatterns.Pattern

	if suppoertedpatternstring, err := json.MarshalIndent(SupportedFilePatternOutput, "", "    "); err == nil {
		return string(suppoertedpatternstring)
	}

	return ""
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func getPlotSupportedFiles() map[string]bool {

	var plothomeDirPath = GetRSHome() + "/plugins/plot_toc_data_provider/plugin_def.xml"
	// Open our xmlFile
	xmlFile, err := os.Open(plothomeDirPath)
	// // if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)
	// we initialize our Users array

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &Plotplugin)

	var plotMap = make(map[string]bool)
	for i := 0; i < len(Plotplugin.DataProvider.SupportedFiles.File); i++ {
		plotMap[Plotplugin.DataProvider.SupportedFiles.File[i].Value] = Plotplugin.DataProvider.SupportedFiles.File[i].IsDefault
	}
	return plotMap
}

func getAnimationSupportedFiles() map[string]bool {

	var plothomeDirPath = GetRSHome() + "/plugins/anim_toc_data_provider/plugin_def.xml"
	// Open our xmlFile
	xmlFile, err := os.Open(plothomeDirPath)
	// // if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)
	// we initialize our Users array

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &Animationplugin)

	var animationMap = make(map[string]bool)
	for i := 0; i < len(Animationplugin.DataProvider.SupportedFiles.File); i++ {
		animationMap[Animationplugin.DataProvider.SupportedFiles.File[i].Value] = Animationplugin.DataProvider.SupportedFiles.File[i].IsDefault
	}
	return animationMap

}

func getRVPSupportedFiles() map[string]bool {

	var plothomeDirPath = GetRSHome() + "/plugins/rvp_toc_data_provider/plugin_def.xml"
	// Open our xmlFile
	xmlFile, err := os.Open(plothomeDirPath)
	// // if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)
	// we initialize our Users array

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &RVPplugin)
	RvpFilesModel = ReadSupportedFilesElement(RVPplugin.DataProvider.SupportedFiles)

	var rvpMap = make(map[string]bool)
	for i := 0; i < len(RVPplugin.DataProvider.SupportedFiles.File); i++ {
		rvpMap[RVPplugin.DataProvider.SupportedFiles.File[i].Pattern] = RVPplugin.DataProvider.SupportedFiles.File[i].IsDefault
	}

	return rvpMap

}

func ReadSupportedFilesElement(supportedFiles datamodel.SupportedFiles) datamodel.SupportedRVPFilesModel {
	var lstFileElement = supportedFiles.File
	var rvpFilesModel datamodel.SupportedRVPFilesModel
	var lstRVPFileModel []datamodel.RVPFileModel

	for i := 0; i < len(lstFileElement); i++ {
		lstRVPFileModel = append(lstRVPFileModel, ReadFileElement(lstFileElement[i]))
	}

	rvpFilesModel.ListRVPFileModel = lstRVPFileModel
	return rvpFilesModel

}

func ReadFileElement(file datamodel.File) datamodel.RVPFileModel {

	var rvpFileModel datamodel.RVPFileModel
	rvpFileModel.IsDefault = file.IsDefault
	rvpFileModel.SupportsDirectPlotOperation = file.SupportsDirectPlotOperation
	rvpFileModel.Pattern = file.Pattern
	rvpFileModel.RvpFileTranslator = parseRVPFileTranslatorModel(file.Translator)
	rvpFileModel.FileParsingStrategiesModel = parseFileParsingStrategies(file.ParsingStrategies)
	return rvpFileModel

}

func parseRVPFileTranslatorModel(translatorElement datamodel.Translator) datamodel.RvpFileTranslator {

	var rvpFileTranslator datamodel.RvpFileTranslator
	if translatorElement.ScriptAbsolutePath != "" {
		rvpFileTranslator.ScriptAbsolutePath = translatorElement.ScriptAbsolutePath
		rvpFileTranslator.TemporaryOutputRVPFileExtension = translatorElement.TemporaryOutputFileExtension
		rvpFileTranslator.ResultFileAbsolutePathArgName = translatorElement.ResultFileAbsolutePathArgName
		rvpFileTranslator.TemporaryRVPFileAbsolutePathArgName = translatorElement.TemporaryFileAbsolutePathArgName

		var scriptparametermap = make(map[string]string)

		for i := 0; i < len(translatorElement.ScriptParameters.ScriptParameter); i++ {
			scriptparametermap[translatorElement.ScriptParameters.ScriptParameter[i].Key] = translatorElement.ScriptParameters.ScriptParameter[i].Value
		}

		rvpFileTranslator.ScriptParameters = scriptparametermap
	}

	return rvpFileTranslator

}

func parseFileParsingStrategies(parsingStrategies datamodel.ParsingStrategies) datamodel.FileParsingStrategiesModel {
	var fileParsingStrategiesModel datamodel.FileParsingStrategiesModel

	var listParsingStrategy = parsingStrategies.ParsingStrategy
	if len(listParsingStrategy) != 0 {
		for i := 0; i < len(listParsingStrategy); i++ {
			fileParsingStrategiesModel.ListFileParsingStrategyModel =
				append(fileParsingStrategiesModel.ListFileParsingStrategyModel, parseFileParsingStrategy(listParsingStrategy[i]))
		}

	}

	return fileParsingStrategiesModel
}

func parseFileParsingStrategy(parsingStrategy datamodel.ParsingStrategy) datamodel.FileParsingStrategyModel {
	var fileParsingStrategyModel datamodel.FileParsingStrategyModel
	fileParsingStrategyModel.Id = parsingStrategy.Id
	fileParsingStrategyModel.ColumnNamesParserModel = parseColumnNames(parsingStrategy.ColumnNames)
	fileParsingStrategyModel.DataPointsParserModel = parseDataPoints(parsingStrategy.DataPoints)
	fileParsingStrategyModel.CommentsParserModel = parseComments(parsingStrategy.Comments)
	return fileParsingStrategyModel

}
func parseColumnNames(columnNames datamodel.ColumnNames) datamodel.ColumnNamesParserModel {

	var columnNamesParserModel datamodel.ColumnNamesParserModel
	columnNamesParserModel.Prefix = columnNames.Prefix
	columnNamesParserModel.Delimiter = columnNames.Delimiter
	columnNamesParserModel.ColumnNamePrefix = columnNames.ColumnNamePrefix

	return columnNamesParserModel

}

func parseDataPoints(dataPoints datamodel.DataPoints) datamodel.DataPointsParserModel {

	var dataPointsParserModel datamodel.DataPointsParserModel
	dataPointsParserModel.Delimiter = dataPoints.Delimiter
	dataPointsParserModel.Prefix = dataPoints.Prefix
	//dataPointsParserModel.NumberLocale = parseLocale(dataPointsParserModel.NumberLocale)

	return dataPointsParserModel

}

func parseLocale(localeElement datamodel.NumberLocale) {

	if localeElement.Language != "" {
		var language = localeElement.Language
		var country = localeElement.Country
		if IsValidString(language) && IsValidString(country) {
			//	loca
		}
	}

	//return locale
}

func parseComments(comments datamodel.Comments) datamodel.CommentsParserModel {
	var commentsParserModel datamodel.CommentsParserModel
	commentsParserModel.Prefix = comments.Prefix
	return commentsParserModel
}

func getRVPDataSupportedFiles() map[string]bool {

	var plothomeDirPath = GetRSHome() + "/plugins/rvp_plot_data_provider/plugin_def.xml"
	// Open our xmlFile
	xmlFile, err := os.Open(plothomeDirPath)
	// // if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)
	// we initialize our Users array

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &RVPDataplugin)

	var rvpdataMap = make(map[string]bool)
	for i := 0; i < len(RVPDataplugin.DataProvider.SupportedFiles.File); i++ {
		if RVPDataplugin.DataProvider.SupportedFiles.File[i].SupportsDirectPlotOperation {
			rvpdataMap[RVPDataplugin.DataProvider.SupportedFiles.File[i].Pattern] = RVPDataplugin.DataProvider.SupportedFiles.File[i].SupportsDirectPlotOperation
		}
	}
	fmt.Println("rvpdataMap", rvpdataMap)
	return rvpdataMap

}

func GetHWComposeConfigDetails() string {

	var isHWConfigured, _ = isHWInstalledandConfigured()
	var isComposeConfigured = isComposeInstalledandConfigured()
	var outputdata = "{\"isHWConfigured\":\"" + strconv.FormatBool(isHWConfigured) + "\",\"isComposeConfigured\":\"" +
		strconv.FormatBool(isComposeConfigured) + "\"}"

	var hwcomposepath = "[{\"WLMSERVER\":" + outputdata + "}]"

	return hwcomposepath
}

func isHWInstalledandConfigured() (bool, error) {
	var isHwInstalledAndConfigured = false

	var HWInstalledDirPath = GetProductInstallationLocation(HYPERWORKS_PRODUCT_ID)

	if HWInstalledDirPath == "" {
		isHwInstalledAndConfigured = false
	}
	fmt.Println("HWLoc: ", HWInstalledDirPath+"/scripts")
	if HWInstalledDirPath != "" {
		dir, _ := os.Stat(HWInstalledDirPath + "/scripts")
		if dir != nil {
			isHwInstalledAndConfigured = true
		} else {
			var sMessage = "Specified HW Product Installation location doesn't exist - " +
				HWInstalledDirPath + "/scripts"
			return false, errors.New(sMessage)
		}
	}

	return isHwInstalledAndConfigured, errors.New("")

}

func isComposeInstalledandConfigured() bool {
	var isComposeInstalledAndConfigured = false

	var ComposeInstalledDirPath = GetProductInstallationLocation(COMPOSE_PRODUCT_ID)

	if ComposeInstalledDirPath == "" {
		isComposeInstalledAndConfigured = false
	}

	if ComposeInstalledDirPath != "" {
		dir, _ := os.Stat(ComposeInstalledDirPath + "/hwx")
		if dir != nil {
			isComposeInstalledAndConfigured = true
		} else {
			var sMessage = "Specified Compose Product Installation location doesn't exist - " +
				ComposeInstalledDirPath + "/hwx"
			log.Fatal(sMessage)
		}
	}

	return isComposeInstalledAndConfigured

}

func GetSupportedSeriesFilePatterns(sToken string) string {

	seriespatternFile = new(datamodel.SeriespatternFile)

	for i := 0; i < len(SiteConfigData.SeriesResultFiles.ResultFile); i++ {

		seriespatternFile.ListSupportedRvsSeriesFilePattern = append(seriespatternFile.ListSupportedRvsSeriesFilePattern,
			datamodel.SeriesPattern{SiteConfigData.SeriesResultFiles.ResultFile[i].SeriesPattern})
		SeriesVsBaseRegEx[SiteConfigData.SeriesResultFiles.ResultFile[i].SeriesPattern] =
			SiteConfigData.SeriesResultFiles.ResultFile[i].BasenamePattern
		SeriesRegexVsWildcard[SiteConfigData.SeriesResultFiles.ResultFile[i].SeriesPattern] =
			SiteConfigData.SeriesResultFiles.ResultFile[i].SeriesWildcardPattern
	}
	fmt.Println(" GetSupportedSeriesFilePatterns SeriesRegexVsWildcard ", SeriesRegexVsWildcard)
	if seriesFilepattern, err := json.MarshalIndent(seriespatternFile, "", "    "); err == nil {
		//	fmt.Println(string(seriesFilepattern))
		return string(seriesFilepattern)
	}

	return ""
}

func GetLastModificationTime(JobState string, JobId string, sPASURL string, FilePath string, sToken string) string {

	url := buildFileListUrl(JobState, JobId, sPASURL)
	postData := buildFilePostData(FilePath, JobId)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(postData))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", sToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	req.Header.Set("Authorization", sToken)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var LastMoifiedoutput lastMoifiedoutput
	json.Unmarshal(body, &LastMoifiedoutput)
	var lastModifiedTime = strconv.FormatInt(LastMoifiedoutput.Data.Files[0].Modified, 10)

	return lastModifiedTime
}

func buildFileListUrl(JobState string, JobId string, sPASURL string) string {

	if JobState == "R" && strings.Contains(sPASURL, PAS_URL_VALUE) {
		sPASURL = strings.Replace(sPASURL, PAS_URL_VALUE, JOB_OPERATION, -1)
	} else {
		sPASURL = sPASURL + REST_SERVICE_URL
	}

	sPASURL = sPASURL + "/files/list"
	return sPASURL
}

func buildFilePostData(sFilePath string, sJobId string) string {

	type pasInput struct {
		Path  string `json:"path"`
		Jobid string `json:"jobid"`
	}

	var data pasInput
	data.Path = sFilePath
	data.Jobid = sJobId

	if outstring, err := json.MarshalIndent(data, "", "    "); err == nil {
		return string(outstring)
	}

	return ""
}

func DownloadMultiFileAsZip(sUrl string, urlParameters string, userAuthToken string, saveFilePath string) string {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", sUrl, bytes.NewBufferString(urlParameters))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", userAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", "application/octet-stream")
	req.Header.Set("charset", "utf-8")
	req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var disPosition = resp.Header.Get("content-disposition")
	var fileName string
	if disPosition != "" {
		var index = strings.Index(disPosition, "filename=")
		if index > 0 {
			fileName = disPosition[index+9:]
		}
	}

	fileName = strings.Replace(fileName, "\"", "", -1)
	if fileName == "" {
		return ""
	}
	saveFilePath = saveFilePath + fileName

	fo, err := os.Create(saveFilePath)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	w := bufio.NewWriter(fo)

	if _, err := w.Write(body); err != nil {
		panic(err)
	}

	return saveFilePath

}
