package common

import (
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var PBSDSDownloadConcurrentRequest pbsDSDownloadConcurrentRequest

type pbsDSDownloadConcurrentRequest struct {
	LstOngoingRequests []DownloadedPBSDSObject `json:"m_lstOngoingRequests"`
}

type DownloadedPBSDSObject struct {
	RemoteFilePath         string `json:"m_remoteFilePath"`
	JobId                  string `json:"m_jobId"`
	LocalFilePath          string `json:"m_localFilePath"`
	IOngoingRequestCounter int    `json:"m_iOngoingRequestCounter"`
}

func ResolvePBSPortDataSource(datasource datamodel.ResourceDataSource, username string, password string) (string, error) {

	log.Println("Entering method resolveDataSource")

	var sFilePath = datasource.FilePath
	var PbsServerData = datasource.PbsServer
	var sPBSServerName = PbsServerData.Server
	var sPortNo = PbsServerData.Port
	var userAuthToken = PbsServerData.AuthorizationToken
	var isPASSecure = PbsServerData.IsSecure
	var pasusername = datasource.PbsServer.UserName
	var paspassword = datasource.PbsServer.UserPassword
	var sPasURL = PbsServerData.PasURL

	if IsWindows() {
		fmt.Println("Windows Server")
	} else {
		log.Println("Checking if user " + username + " has read permission on file ")
		var arrCmd []string
		if Is32BitOS() {
			arrCmd = append(arrCmd, GetRSHome()+"/bin/linux32/CheckPermission.sh")
		} else {
			fmt.Println("Linux 64 Server")
			arrCmd = append(arrCmd, GetRSHome()+"/bin/linux64/CheckPermission.sh")
		}
		arrCmd = append(arrCmd, sFilePath)
		fmt.Println("arrCmd", arrCmd)
		var iExitCode = RunCommand(arrCmd, username, password)
		fmt.Println("iExitCode ", iExitCode)
		if iExitCode == 0 {
			log.Println("user " + username + " has read permission on file " + sFilePath)
			log.Println(
				"Adding file path as custom entry within datasource to be used by subsequent request")
			datasource.Custom.Any = append(datasource.Custom.Any, sFilePath)
			return sFilePath, nil
		} else if iExitCode == 1 {
			log.Println("User " + username + " does not have direct permission on file " + sFilePath + ", lets download it")
			var sJobId = datasource.PbsServer.JobId
			var sJobStatus = datasource.PbsServer.JobStatus
			var bIsJobRunning = false

			if PBS_JOB_EXIT_STATE == sJobStatus {
				var sMessage = "Job: " + sJobId + " in exit state."
				log.Println(sMessage)
				return "", &exception.RVSError{
					Errordetails: "",
					Errorcode:    "3004",
					Errortype:    "TYPE_AUTH_FAILED",
				}
			} else if PBS_JOB_RUNNING_STATE == sJobStatus {
				bIsJobRunning = true
			}

			var sDownloadedFilePath = downloadFileFromPBSServer(sFilePath,
				sJobId, bIsJobRunning, sPBSServerName, sPortNo, isPASSecure,
				username, password, pasusername, paspassword, datasource, userAuthToken, sPasURL)

			datasource.Custom.Any = append(datasource.Custom.Any, sDownloadedFilePath)
			return sDownloadedFilePath, nil

		}
	}

	return "", nil

}

/**
     * This method checks if any similar download is going on.
     * If it finds such a request, it pools the current request.
     * If not, it downloads file from pbs server.
     *
	 **/

func downloadFileFromPBSServer(sFilePath string, sJobId string, bIsJobRunning bool,
	sServerName string, sPortNo string, isPASSecure bool,
	username string, password string,
	pbsusername string, pbspassword string, datasource datamodel.ResourceDataSource, userAuthToken string, sPASUrl string) string {

	log.Println("Entering method downloadFileFromPBSServer")
	var startTime = time.Now()
	log.Println("Search existing object used for downloading from download concurrent request manager")
	log.Println("This helps in performance in case more than one user has asked for same file from same job")

	var downloadedPbsObject = GetDownloadedPASFileObject(sFilePath, sJobId)
	//downloadedFileObject.Lock()
	var fileDownloaded string

	if downloadedPbsObject.LocalFilePath == "" {
		log.Println("First download request, going to download the file")

		if datasource.SeriesFile {
			log.Println("Downloading series file...")
			fileDownloaded = downloadPBSSeriesFileOnLinux(sFilePath, sJobId, bIsJobRunning,
				sServerName, sPortNo, isPASSecure,
				username, password, pbsusername, pbspassword,
				datasource, userAuthToken, sPASUrl)
		} else {
			log.Println("Downloading non series file...")
			fileDownloaded = downloadFileOnLinux(sFilePath, sJobId, bIsJobRunning, sServerName,
				sPortNo, isPASSecure, username, password, pbsusername, pbspassword,
				datasource, userAuthToken, sPASUrl)
			// set the response on file download object for other waiting threads
			log.Println("Setting downloaded file absolute path in downloadedFileObject to be used by other thread")
		}
		downloadedPbsObject.LocalFilePath = fileDownloaded
	}

	//downloadedFileObject.Unlock()
	var endTime = time.Now()
	log.Println("File Download Time ", endTime.Sub(startTime))
	return fileDownloaded

}

/**
     * This method finds whether any similar request is being computed in the memory
     * or not. If it finds a similar request, it returns the same file download object
     * otherwise it builds an object for the file download request and returns it.
     *
	 **/
func GetDownloadedPASFileObject(sRemoteFilePath string, sJobId string) DownloadedPBSDSObject {
	log.Println("Entering method DownloadedPBSDSObject")
	var sFilePath = GetPlatformIndependentFilePath(sRemoteFilePath, false)
	log.Println("Looking for existing request for file path " + sRemoteFilePath + " and job id " + sJobId)
	var downloadedPbsObject = getExistingPBSRequest(sFilePath, sJobId)
	if downloadedPbsObject.RemoteFilePath == "" {
		log.Println("No existing request found, create new DownloadedFilePortDSObject")
		// No existing similar request - so build it and add it to the ongoing req list
		var downloadedPbsObjectNew = DownloadedPBSDSObject{
			RemoteFilePath:         sFilePath,
			JobId:                  sJobId,
			IOngoingRequestCounter: 1,
		}

		PBSDSDownloadConcurrentRequest.LstOngoingRequests = append(PBSDSDownloadConcurrentRequest.LstOngoingRequests, downloadedPbsObjectNew)
		return downloadedPbsObjectNew
	} else {
		log.Println("Increment the counter for DownloadedFilePortDSObject")
		downloadedPbsObject.IOngoingRequestCounter = downloadedPbsObject.IOngoingRequestCounter + 1
		return downloadedPbsObject
	}
}

func getExistingPBSRequest(sRemoteFilePath string, jobId string) DownloadedPBSDSObject {
	log.Println("Entering method getExistingRequest")

	for i := 0; i < len(PBSDSDownloadConcurrentRequest.LstOngoingRequests); i++ {
		if PBSDSDownloadConcurrentRequest.LstOngoingRequests[i].RemoteFilePath == sRemoteFilePath &&
			PBSDSDownloadConcurrentRequest.LstOngoingRequests[i].JobId == jobId {
			return PBSDSDownloadConcurrentRequest.LstOngoingRequests[i]
		}
	}
	log.Println("Found no existing request")
	return DownloadedPBSDSObject{}
}

func downloadPBSSeriesFileOnLinux(sFilePath string, sJobId string, bIsJobRunning bool,
	sServerName string, sPortNo string, isPASSecure bool,
	username string, password string, pbsusername string, pbspassword string,
	dataSource datamodel.ResourceDataSource, userAuthToken string, sPASUrl string) string {
	log.Println("Entering method downloadPBSSeriesFileOnLinux")

	var lstChangedFiles []string
	log.Println("Retrieving last modified time for all series files present")
	var mapCurrentFileVsModTime = getPBSLastModificationTimeForParentDir(dataSource, sFilePath, bIsJobRunning, userAuthToken, sPASUrl)

	//TODO ad cache code

	//Download first time
	lstChangedFiles = getAllPbsFilesList(mapCurrentFileVsModTime)
	var fileDownloaded = readSeriesPBSFileFromPBSServerUserToken(sFilePath, sJobId, bIsJobRunning, sServerName, sPortNo,
		isPASSecure, username, password, pbsusername, pbspassword,
		dataSource, userAuthToken, lstChangedFiles, "", sPASUrl)
	fmt.Println(fileDownloaded)
	return fileDownloaded

}

func getPBSLastModificationTimeForParentDir(dataSource datamodel.ResourceDataSource, sFilePath string, bIsJobRunning bool,
	userAuthToken string, sPASUrl string) map[string]int64 {
	log.Println("Entering method getPBSLastModificationTimeForParentDir")
	var sParentDirPath = GetPlatformIndependentFilePath(filepath.Dir(dataSource.FilePath), false)
	var mapFileVsModTime = make(map[string]int64)
	var pbsServer = dataSource.PbsServer
	var sJobId = ""
	log.Println("Getting FileOperations port")

	log.Println("User:" + pbsServer.UserName + ": Checking file access permission: " + sParentDirPath)
	var sJobStatus = pbsServer.JobStatus
	if PBS_JOB_RUNNING_STATE == sJobStatus {
		sJobId = pbsServer.JobId
	}

	log.Println("Calling fileList operation on job id" + sJobId + " and parent dir path " + sParentDirPath)

	var fileListResult = getPbsFileListWLM(sJobId, bIsJobRunning, sParentDirPath, userAuthToken, sPASUrl)

	if len(fileListResult) != 0 {
		log.Println("Parent dir may contain other files than series files. lets filter them")
		var lstFilteredFileData = filterFileData(fileListResult, sFilePath)

		log.Println("Adding last modified time in a map for all the filtered files")
		for i := 0; i < len(lstFilteredFileData); i++ {
			var sFileName = lstFilteredFileData[i].Filename
			var lModTime = lstFilteredFileData[i].Modified
			mapFileVsModTime[sFileName] = lModTime
		}

		return mapFileVsModTime
	}
	return mapFileVsModTime
}

func getPbsFileListWLM(sJobId string, bIsJobRunning bool, sParentDirPath string, userAuthToken string, sPASUrl string) []FileModel {
	var url = buildPbsFileListUrl(sPASUrl, bIsJobRunning)
	var sPostData = buildPbsPostData(sParentDirPath, sJobId)
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

	return parsePbsFileListJson(response)
}

func buildPbsFileListUrl(sPasURL string, bIsJobRunning bool) string {
	if bIsJobRunning && strings.Contains(sPasURL, PAS_URL_VALUE) {
		sPasURL = strings.Replace(sPasURL, PAS_URL_VALUE, JOB_OPERATION, -1)
	} else {
		sPasURL = sPasURL + REST_SERVICE_URL
	}
	return (sPasURL + "/files/list")
}

func buildPbsPostData(sFilePath string, sJobId string) string {
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

func parsePbsFileListJson(response []byte) []FileModel {

	var data Modiedfilelist
	json.Unmarshal(response, &data)
	return data.Data.FileModel

}
func getAllPbsFilesList(mapCurrentFileVsModTime map[string]int64) []string {

	var lstChangedFiles []string

	for sFileName, _ := range mapCurrentFileVsModTime {
		if mapCurrentFileVsModTime[sFileName] != 0 {
			lstChangedFiles = append(lstChangedFiles, sFileName)
		}
	}

	return lstChangedFiles
}

func readSeriesPBSFileFromPBSServerUserToken(sFilePath string, sJobId string, bIsJobRunning bool, sServerName string, sPortNo string,
	isPASSecure bool, username string, password string, pbsusername string, pbspassword string,
	dataSource datamodel.ResourceDataSource, userAuthToken string, filesToDownload []string, fileToWrite string, sPASUrl string) string {

	log.Println("Downloading file from WLM using WLM API")
	if fileToWrite == "" {
		fileToWrite = getDownloadFilePath(dataSource, username, password, sFilePath)
	}

	var sUrl = buildPbsMultiDownloadUrl(bIsJobRunning, sPASUrl)
	var urlParameters = buildPbsURLParametres(filesToDownload, sFilePath, sJobId)
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
	log.Println("File Service Download Time ", time.Since(wirteTime))
	// Get last modification time
	log.Println("Getting last modification time of datasource")
	var JobState = ""
	if bIsJobRunning {
		JobState = "R"
	}
	var lastModTime = GetLastModificationTime(JobState, sJobId, sPASUrl, sFilePath, userAuthToken)
	fmt.Println("lastModTime", lastModTime)
	// Set original last modification time
	log.Println("Setting last modification time received from datasource on the local temp file ", fileToWrite)
	var timeInMilllis, _ = strconv.ParseInt(lastModTime, 10, 64)

	err := os.Chtimes(fileToWrite, time.Now(), time.UnixMilli(timeInMilllis))
	if err != nil {
		fmt.Println(err)
	}
	return fileToWrite

}

func buildPbsMultiDownloadUrl(bIsJobRunning bool, sPASUrl string) string {

	if bIsJobRunning && strings.Contains(sPASUrl, PAS_URL_VALUE) {
		sPASUrl = strings.Replace(sPASUrl, PAS_URL_VALUE, JOB_OPERATION, -1)
	} else {
		sPASUrl = sPASUrl + REST_SERVICE_URL
	}
	sPASUrl = sPASUrl + "/files/downloadMulti"

	return sPASUrl

}

func buildPbsURLParametres(lstFilePath []string, sFilePath string, sJobID string) string {
	var urlParameters = ""
	var path = GetDirPath(sFilePath)
	for i := 0; i < len(lstFilePath); i++ {
		if i == 0 {
			urlParameters = urlParameters + "paths=" + path + lstFilePath[i]
		} else {
			urlParameters = urlParameters + "&paths=" + path + lstFilePath[i]
		}
	}
	if sJobID != "" {
		urlParameters = urlParameters + "&" + "jobid=" + sJobID
	}
	return urlParameters
}

func downloadFileOnLinux(sFilePath string, sJobId string, bIsJobRunning bool, sServerName string,
	sPortNo string, isPASSecure bool, username string, password string, pbsusername string, pbspassword string,
	dataSource datamodel.ResourceDataSource, userAuthToken string, sPASUrl string) string {

	log.Println("Entering downloadFileOnLinux downloadFile")
	//TODO Cache

	log.Println("Copying files via SCP/RVP failed. Falling back to AIF file copy...")
	var fileDownloaded = readFileFromPBSServerUserToken(sFilePath,
		sJobId, bIsJobRunning, sServerName, sPortNo,
		isPASSecure, username, password, pbsusername, pbspassword, dataSource, userAuthToken, sPASUrl)

	//TODO CACHING
	return fileDownloaded
}

func readFileFromPBSServerUserToken(sFilePath string,
	sJobId string, bIsJobRunning bool, sServerName string, sPortNo string,
	isPASSecure bool, username string, password string, pbsusername string, pbspassword string,
	dataSource datamodel.ResourceDataSource, userAuthToken string, sPASUrl string) string {
	log.Println("Downloading file from WLM using WLM API")
	var fileToWrite = getPbsDownloadFilePath(dataSource, username, password, sFilePath)
	log.Println("sJobId:" + sJobId + "bIsJobRunning" + strconv.FormatBool(bIsJobRunning))
	var sUrl = buildDownloadUrl(sServerName, sPortNo, isPASSecure, sFilePath, sJobId, bIsJobRunning, sPASUrl)
	log.Println("downloadUrl:" + sUrl)

	var data = downloadPBSFile(sUrl, userAuthToken)

	f, writeerr := os.Create(fileToWrite)

	if writeerr != nil {
		log.Fatal(writeerr)
	}

	defer f.Close()

	_, dataerr2 := f.WriteString(data)

	if dataerr2 != nil {
		log.Fatal(dataerr2)
	}
	// Get last modification time
	log.Println("Getting last modification time of datasource")
	var JobState = ""
	if bIsJobRunning {
		JobState = "R"
	}
	var timeInMilllisString = GetLastModificationTime(JobState, sJobId, sPASUrl, sFilePath, userAuthToken)
	var timeInMilllis, _ = strconv.ParseInt(timeInMilllisString, 10, 64)

	err := os.Chtimes(fileToWrite, time.Now(), time.UnixMilli(timeInMilllis))
	if err != nil {
		fmt.Println(err)
	}
	return fileToWrite

}

func getPbsDownloadFilePath(dataSource datamodel.ResourceDataSource, username string, passowrd string, sFilePath string) string {

	var sFilePathNew = GetPlatformIndependentFilePath(sFilePath, false)

	var parentFolder = AllocateUniqueFolder(SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+RM_DOWNLOADS, "DOWNLOAD")

	var fileToWrite = AllocateFileWithGlobalPermission(GetFileName(sFilePathNew), parentFolder)
	log.Println("Created temp file " + fileToWrite + " to store the pbs data source")
	return fileToWrite
}

func buildDownloadUrl(sServerName string, sPortNo string, isSecure bool, serverSideFilePath string, sJobId string,
	bIsJobRunning bool, sPASUrl string) string {
	if bIsJobRunning && strings.Contains(sPASUrl, PAS_URL_VALUE) {
		sPASUrl = strings.Replace(sPASUrl, PAS_URL_VALUE, JOB_OPERATION, -1)
	} else {
		sPASUrl = sPASUrl + REST_SERVICE_URL
	}
	sPASUrl = sPASUrl + "/files/download"
	log.Println("downloadUrl:" + sPASUrl)
	return sPASUrl
}

func downloadPBSFile(sUrl string, userAuthToken string) string {

	fmt.Println("urlstring:", sUrl)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("GET", sUrl, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", userAuthToken)

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
