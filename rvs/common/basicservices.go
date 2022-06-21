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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var Plotplugin datamodel.Plugin
var Animationplugin datamodel.Plugin
var RVPplugin datamodel.Plugin
var RVPDataplugin datamodel.Plugin
var SupportedFilePatternOutput *datamodel.SupportedFilePatternOutput
var seriespatternFile *datamodel.SeriespatternFile

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
		//fmt.Println(string(xmlstring))
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

	var rvpMap = make(map[string]bool)
	for i := 0; i < len(RVPplugin.DataProvider.SupportedFiles.File); i++ {
		rvpMap[RVPplugin.DataProvider.SupportedFiles.File[i].Pattern] = RVPplugin.DataProvider.SupportedFiles.File[i].IsDefault
	}
	fmt.Println("rvpMap", rvpMap)
	return rvpMap

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
	var postData = []byte(urlParameters)
	var postDataLength = len(postData)

	req.Header.Set("Authorization", userAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("accept", "application/json")
	req.Header.Set("charset", "utf-8")
	req.Header.Set("Content-Length", strconv.Itoa(postDataLength))
	//req.Header.Set("content-disposition", "filename="+GetFileName(saveFilePath))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("res", resp)
	var disPosition = resp.Header.Get("content-disposition")
	//fmt.Println("disPosition", disPosition)
	var fileName string
	if disPosition != "" {
		var index = strings.Index(disPosition, "filename=")
		if index > 0 {
			fileName = disPosition[index+9:]
		}
	}

	fileName = strings.Replace(fileName, "\"", "", -1)
	saveFilePath = saveFilePath + fileName
	fmt.Println("fileName", fileName)
	fmt.Println("saveFilePath", saveFilePath)

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
