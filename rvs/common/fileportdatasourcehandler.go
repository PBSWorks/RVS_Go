package common

import (
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"archive/zip"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Modiedfilelist struct {
	Success  bool   `json:"success"`
	Data     Data   `json:"data"`
	ExitCode string `json:"exitCode"`
}

type Data struct {
	FileModel []FileModel `json:"files"`
}

type FileModel struct {
	Created     int64  `json:"created"`
	Filename    string `json:"filename"`
	Modified    int64  `json:"modified"`
	Owner       string `json:"owner"`
	Size        string `json:"size"`
	Type        string `json:"type"`
	IsWritable  string `json:"isWritable"`
	AbsPath     string `json:"absPath"`
	HasChildren string `json:"hasChildren"`
	IsReadable  string `json:"isReadable"`
	IsHidden    string `json:"isHidden"`
}

var FilePortDownloadConcurrentReq filePortDownloadConcurrentReq

type filePortDownloadConcurrentReq struct {
	LstOngoingRequests []DownloadedFilePortDSObject `json:"m_lstOngoingRequests"`
}

type DownloadedFilePortDSObject struct {
	RemoteFilePath         string `json:"m_remoteFilePath"`
	ServerName             string `json:"m_serverName"`
	LocalFilePath          string `json:"m_localFilePath"`
	IOngoingRequestCounter int    `json:"m_iOngoingRequestCounter"`
}

func ResolveFilePortDataSource(datasource datamodel.ResourceDataSource, username string, password string) (string, error) {

	var sFilePath = datasource.FilePath
	var sFilePortServerName = datasource.FilePortServer.Port
	var fileportusername = datasource.FilePortServer.UserName
	var fileportpassword = datasource.FilePortServer.UserPassword
	var userAuthToken = datasource.FilePortServer.AuthorizationToken
	// // If the user doesn't have file access permissions
	var sPasURL = datasource.FilePortServer.PasUrl

	if IsWindows() {
		fmt.Println("Windows Server")
	} else {
		log.Println("Checking if user " + username + " has read permission on file ")
		var arrCmd []string
		if Is32BitOS() {
			arrCmd = append(arrCmd, GetRSHome()+"/bin/linux32/CheckPermission.sh")
		} else {
			arrCmd = append(arrCmd, GetRSHome()+"/bin/linux64/CheckPermission.sh")
		}
		arrCmd = append(arrCmd, sFilePath)
		var iExitCode = RunCommand(arrCmd, username, password)

		if iExitCode == 0 {
			log.Println("user " + username + " has read permission on file " + sFilePath)
			log.Println(
				"Adding file path as custom entry within datasource to be used by subsequent request")
			datasource.Custom.Any = append(datasource.Custom.Any, sFilePath)
			return sFilePath, nil
		} else if iExitCode == 1 {
			log.Println("user " + username + " does not have read permission on file " + sFilePath + ", need to download file")
			// Download file
			var sDownloadedFilePath = downloadFileFromFilePortServer(sFilePath,
				sFilePortServerName, username, password, fileportusername, fileportpassword,
				datasource, userAuthToken, sPasURL)
			log.Println(
				"Adding file path as custom entry within datasource to be used by subsequent request")
			return sDownloadedFilePath, nil
		} else if iExitCode == 2 {
			var sMessage = "Datasource file: " + sFilePath + " is not accessible to the " + "user: " + fileportusername
			log.Println(sMessage)
			return "", &exception.RVSError{
				Errordetails: "",
				Errorcode:    "3004",
				Errortype:    "TYPE_AUTH_FAILED",
			}
		} else {
			var sMessage = "Datasource file: " + sFilePath + " is not accessible to the " + "user: " + fileportusername
			log.Println(sMessage)
			return "", &exception.RVSError{
				Errordetails: "",
				Errorcode:    "3004",
				Errortype:    "TYPE_AUTH_FAILED",
			}
		}

	}
	return "", nil
}

func downloadFileFromFilePortServer(sFilePath string,
	sFilePortServerName string, username string, password string, fileportusername string, fileportpassword string,
	dataSource datamodel.ResourceDataSource, userAuthToken string, sPasURL string) string {
	log.Println("Entering method downloadFileFromFilePortServer")
	log.Println("Search for any same parallel request made using fileDownloadConcurrentRequestManager")
	var downloadedFileObject DownloadedFilePortDSObject
	downloadedFileObject = GetDownloadedFileObject(sFilePath, sFilePortServerName)

	//downloadedFileObject.Lock()
	var fileDownloaded string
	if downloadedFileObject.LocalFilePath == "" {
		log.Println("First download request, going to download the file")

		if dataSource.SeriesFile {
			log.Println("Downloading series file...")
			fileDownloaded = downloadSeriesFileOnLinux(sFilePath, sFilePortServerName,
				username, password, fileportusername, fileportpassword, dataSource, userAuthToken, sPasURL)
		} else {
			fileDownloaded = downloadFile(sFilePath, sFilePortServerName,
				username, password, fileportusername, fileportpassword, dataSource, userAuthToken, sPasURL)
			// set the response on file download object for other waiting threads
			log.Println("Setting downloaded file absolute path in downloadedFileObject to be used by other thread")
		}
	}

	//downloadedFileObject.Unlock()

	return fileDownloaded
}

func downloadSeriesFileOnLinux(sFilePath string, sFilePortServerName string,
	username string, password string, fileportusername string, fileportpassword string, dataSource datamodel.ResourceDataSource,
	userAuthToken string, sPasURL string) string {
	log.Println("Entering method downloadSeriesFileOnLinux")

	var lstChangedFiles []string
	log.Println("Retrieving last modified time for all series files present")
	var mapCurrentFileVsModTime = getLastModificationTimeForParentDir(dataSource, sFilePath, userAuthToken, sPasURL)

	//TODO ad cache code

	//Download first time
	lstChangedFiles = getAllFilesList(mapCurrentFileVsModTime)
	var fileDownloaded = readSeriesFileFromPBSServerUserToken(sFilePath, sFilePortServerName,
		username, password, dataSource, userAuthToken, lstChangedFiles, "", sPasURL)
	fmt.Println(fileDownloaded)
	return fileDownloaded

}

func getLastModificationTimeForParentDir(dataSource datamodel.ResourceDataSource,
	sFilePath string, userAuthToken string, sPasURL string) map[string]int64 {
	log.Println("Entering method getLastModificationTimeForParentDir")
	var sParentDirPath = GetPlatformIndependentFilePath(filepath.Dir(dataSource.FilePath), false)
	var mapFileVsModTime = make(map[string]int64)
	log.Println("Getting FileOperations port")

	var fileListResult = getFileListWLM(sParentDirPath, userAuthToken, sPasURL)
	if len(fileListResult) != 0 {
		log.Println("Parent dir may contain other files than series files. lets filter them")
		var lstFilteredFileData = filterFileData(fileListResult, sFilePath)
		log.Println("Adding last modified time in a map for all the filtered files")
		for i := 0; i < len(lstFilteredFileData); i++ {
			var sFileName = lstFilteredFileData[i].Filename
			var lModTime = lstFilteredFileData[i].Modified
			mapFileVsModTime[sFileName] = lModTime
		}
		//log.Println("Time taken by fileList: " + (System.currentTimeMillis()-start)*1.0/1000)
		return mapFileVsModTime
	}
	return mapFileVsModTime
}

func getFileListWLM(sParentDirPath string, userAuthToken string, sPasURL string) []FileModel {
	var url = buildWlmFileListUrl(sPasURL)
	var sPostData = buildWlmPostData(sParentDirPath)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(sPostData))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", userAuthToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return parseFileListJson(response)

}

func buildWlmFileListUrl(sPasURL string) string {
	//URL url = null;
	sPasURL = sPasURL + "/restservice/files/list"
	fmt.Println("sPasURL", sPasURL)
	return sPasURL
}

func buildWlmPostData(sFilePath string) string {
	type pasInput struct {
		Path string `json:"path"`
	}

	var data pasInput
	data.Path = sFilePath

	if outstring, err := json.MarshalIndent(data, "", "    "); err == nil {
		return string(outstring)
	}

	return ""
}

func parseFileListJson(response []byte) []FileModel {

	var data Modiedfilelist
	json.Unmarshal(response, &data)

	return data.Data.FileModel

}

func filterFileData(lstFileData []FileModel, sFilePath string) []FileModel {
	log.Println("Entering method filterFileData")
	var lstFilteredFileData []FileModel
	var sRegex = getSeriesRegexPatternForFileName(GetFileName(sFilePath))

	re, _ := regexp.Compile(sRegex)
	for i := 0; i < len(lstFileData); i++ {
		match := re.FindString(lstFileData[i].Filename)
		if match != "" {
			log.Println("Adding filtered file " + lstFileData[i].Filename)
			lstFilteredFileData = append(lstFilteredFileData, lstFileData[i])
		}
	}
	return lstFilteredFileData
}

func getSeriesRegexPatternForFileName(sFileName string) string {

	for key, _ := range SeriesRegexVsWildcard {
		re, _ := regexp.Compile(key)
		match := re.FindString(sFileName)
		if match != "" {
			return key
		}
	}
	return ""

}

func getAllFilesList(mapCurrentFileVsModTime map[string]int64) []string {

	var lstChangedFiles []string

	for sFileName, _ := range mapCurrentFileVsModTime {
		if mapCurrentFileVsModTime[sFileName] != 0 {
			lstChangedFiles = append(lstChangedFiles, sFileName)
		}
	}

	return lstChangedFiles
}

func readSeriesFileFromPBSServerUserToken(sFilePath string, sFilePortServerName string,
	username string, password string, dataSource datamodel.ResourceDataSource, userAuthToken string,
	filesToDownload []string, fileToWrite string, sPasURL string) string {

	log.Println("Downloading file from WLM using WLM API")
	if fileToWrite == "" {
		fileToWrite = getDownloadFilePath(dataSource, username, password, sFilePath)
	}

	var sUrl = buildMultiDownloadUrl(sFilePortServerName, sFilePath, sPasURL)
	var urlParameters = buildURLParametres(filesToDownload, sFilePath)

	fmt.Println("sUrl", sUrl)
	fmt.Println("urlParameters", urlParameters)
	var fileDownloadStartTime = time.Now()
	var zipFile = DownloadMultiFileAsZip(sUrl, urlParameters, userAuthToken, GetDirPath(fileToWrite))
	var fileDownloadEndTime = time.Now()
	log.Println("File Service Download Time ", fileDownloadEndTime.Sub(fileDownloadStartTime))
	var wirteTime = time.Now()
	if len(filesToDownload) != 1 {
		log.Println("Downloading multifile as zip ")
		_, err := Unzip(zipFile, GetDirPath(fileToWrite))
		if err != nil {
			log.Fatal(err)
		}
		e := os.Remove(zipFile)
		if e != nil {
			log.Fatal(e)
		}
	}
	log.Println("File Service Download Time ", time.Now().Sub(wirteTime))
	// Get last modification time
	log.Println("Getting last modification time of datasource")
	var lastModTime = GetLastModificationTime("", "", dataSource.FilePortServer.PasUrl,
		sFilePath, userAuthToken)
	// Set original last modification time
	log.Println("Setting last modification time received from datasource on the local temp file ", fileToWrite)
	var timeInMilllis, _ = strconv.ParseInt(lastModTime, 10, 64)

	err := os.Chtimes(fileToWrite, time.Now(), time.UnixMilli(timeInMilllis))
	if err != nil {
		fmt.Println(err)
	}
	return fileToWrite

}

func getDownloadFilePath(dataSource datamodel.ResourceDataSource, username string, password string, sFilePath string) string {
	var sFilePathNew = GetPlatformIndependentFilePath(sFilePath, false)

	var parentFolder = AllocateUniqueFolder(SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+RM_DOWNLOADS, "DOWNLOAD")

	var fileToWrite = AllocateFileWithGlobalPermission(GetFileName(sFilePathNew), parentFolder)
	log.Println("Created temp file " + fileToWrite + " to store the pbs data source")
	return fileToWrite

}
func buildMultiDownloadUrl(servername string, sFilePath string, sPasURL string) string {
	sPasURL = sPasURL + "/restservice/files/downloadMulti"
	return sPasURL

}
func buildURLParametres(lstFilePath []string, sFilePath string) string {
	var urlParameters = ""
	var path = GetDirPath(sFilePath)
	for i := 0; i < len(lstFilePath); i++ {
		if i == 0 {
			urlParameters = urlParameters + "paths=" + path + lstFilePath[i]
		} else {
			urlParameters = urlParameters + "&paths=" + path + lstFilePath[i]
		}
	}
	return urlParameters
}

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func downloadFile(sFilePath string, sFilePortServerName string,
	username string, password string, fileportusername string, fileportpassword string, dataSource datamodel.ResourceDataSource,
	userAuthToken string, sPasURL string) string {

	log.Println("Entering method downloadFile")
	//TODO Cache

	log.Println("Copying files via SCP/RVP failed. Falling back to AIF file copy...")
	var fileDownloaded = readFileFromFilePortServerUserToken(sFilePath, username, password, userAuthToken, sPasURL)

	//TODO CACHING
	return fileDownloaded
}

func readFileFromFilePortServerUserToken(sFilePath string, username string, password string,
	userAuthToken string, sPasURL string) string {
	log.Println("Entering method readFileFromFilePortServer")

	sFilePath = GetPlatformIndependentFilePath(sFilePath, false)
	var parentFolder = AllocateUniqueFolder(SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+RM_DOWNLOADS, "DOWNLOAD")
	var fileToWrite = AllocateFileWithGlobalPermission(GetFileName(sFilePath), parentFolder)
	log.Println("Created temp file to store the datasource " + fileToWrite)

	var data = DownloadFileWLM(sPasURL, "", "", sFilePath, userAuthToken)

	f, writeerr := os.Create(fileToWrite)

	if writeerr != nil {
		log.Fatal(writeerr)
	}

	defer f.Close()

	_, dataerr2 := f.WriteString(data)

	if dataerr2 != nil {
		log.Fatal(dataerr2)
	}

	log.Println("Get last modification time for file " + sFilePath)
	var timeInMilllisString = GetLastModificationTime("", "", sPasURL, sFilePath, userAuthToken)
	var timeInMilllis, _ = strconv.ParseInt(timeInMilllisString, 10, 64)

	err := os.Chtimes(fileToWrite, time.Now(), time.UnixMilli(timeInMilllis))
	if err != nil {
		fmt.Println(err)
	}

	return fileToWrite

}

func GetDownloadedFileObject(sRemoteFilePath string, sFilePortServerName string) DownloadedFilePortDSObject {
	log.Println("Entering method getDownloadedFileObject")
	var sFilePath = GetPlatformIndependentFilePath(sRemoteFilePath, false)
	log.Println("Search for any existing request with server name " + sFilePortServerName + " and file path " + sFilePath)
	var downloadedFileObject = getExistingRequest(sFilePath, sFilePortServerName)
	if downloadedFileObject.RemoteFilePath == "" {
		log.Println("No existing request found, create new DownloadedFilePortDSObject")
		// No existing similar request - so build it and add it to the ongoing req list
		var downloadedFileObjectNew = DownloadedFilePortDSObject{
			RemoteFilePath: sFilePath,
			ServerName:     sFilePortServerName,
		}
		FilePortDownloadConcurrentReq.LstOngoingRequests = append(FilePortDownloadConcurrentReq.LstOngoingRequests, downloadedFileObjectNew)
		return downloadedFileObjectNew
	} else {
		log.Println("Increment the counter for DownloadedFilePortDSObject")
		downloadedFileObject.IOngoingRequestCounter = downloadedFileObject.IOngoingRequestCounter + 1
		return downloadedFileObject
	}

}

func getExistingRequest(sRemoteFilePath string, sServerName string) DownloadedFilePortDSObject {
	log.Println("Entering method getExistingRequest")

	for i := 0; i < len(FilePortDownloadConcurrentReq.LstOngoingRequests); i++ {
		if FilePortDownloadConcurrentReq.LstOngoingRequests[i].RemoteFilePath == sRemoteFilePath &&
			FilePortDownloadConcurrentReq.LstOngoingRequests[i].ServerName == sServerName {
			log.Println("Found existing request with server name " + sServerName + " and file path " + sRemoteFilePath)
			return FilePortDownloadConcurrentReq.LstOngoingRequests[i]
		}
	}
	log.Println("Found no existing request")
	return DownloadedFilePortDSObject{}
}
