package common

import (
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"altair/rvs/utils"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	l "altair/rvs/globlog"
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

	l.Log().Info("Entering method resolveDataSource")

	var sFilePath = datasource.FilePath
	var PbsServerData = datasource.PbsServer
	var sPBSServerName = PbsServerData.Server
	var sPortNo = PbsServerData.Port
	var userAuthToken = PbsServerData.AuthorizationToken
	var isPASSecure = PbsServerData.IsSecure
	var pasusername = datasource.PbsServer.UserName
	var paspassword = datasource.PbsServer.UserPassword
	var sPasURL = PbsServerData.PasURL

	if utils.IsWindows() {
		l.Log().Info("Windows Server")
	} else {
		l.Log().Info("Checking if user " + username + " has read permission on file ")
		var arrCmd []string
		if utils.Is32BitOS() {
			arrCmd = append(arrCmd, utils.GetRSHome()+"/bin/linux32/CheckPermission.sh")
		} else {
			arrCmd = append(arrCmd, utils.GetRSHome()+"/bin/linux64/CheckPermission.sh")
		}
		arrCmd = append(arrCmd, sFilePath)
		var iExitCode = RunCommand(arrCmd, username, password)
		if iExitCode == 0 {
			l.Log().Info("user " + username + " has read permission on file " + sFilePath)
			l.Log().Info(
				"Adding file path as custom entry within datasource to be used by subsequent request")
			datasource.Custom.Any = append(datasource.Custom.Any, sFilePath)
			return sFilePath, nil
		} else if iExitCode == 1 {
			l.Log().Info("User " + username + " does not have direct permission on file " + sFilePath + ", lets download it")
			var sJobId = datasource.PbsServer.JobId
			var sJobStatus = datasource.PbsServer.JobStatus
			var bIsJobRunning = false

			if utils.PBS_JOB_EXIT_STATE == sJobStatus {
				var sMessage = "Job: " + sJobId + " in exit state."
				l.Log().Error(sMessage)
				return "", &exception.RVSError{
					Errordetails: sMessage,
					Errorcode:    "3004",
					Errortype:    "TYPE_AUTH_FAILED",
				}
			} else if utils.PBS_JOB_RUNNING_STATE == sJobStatus {
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

	l.Log().Info("Entering method downloadFileFromPBSServer")
	var startTime = time.Now()
	l.Log().Info("Search existing object used for downloading from download concurrent request manager")
	l.Log().Info("This helps in performance in case more than one user has asked for same file from same job")

	var downloadedPbsObject = GetDownloadedPASFileObject(sFilePath, sJobId)
	//downloadedFileObject.Lock()
	var fileDownloaded string

	if downloadedPbsObject.LocalFilePath == "" {
		l.Log().Info("First download request, going to download the file")

		if datasource.SeriesFile {
			l.Log().Info("Downloading series file...")
			fileDownloaded = downloadPBSSeriesFileOnLinux(sFilePath, sJobId, bIsJobRunning,
				sServerName, sPortNo, isPASSecure,
				username, password, pbsusername, pbspassword,
				datasource, userAuthToken, sPASUrl)
		} else {
			l.Log().Info("Downloading non series file...")
			fileDownloaded = downloadFileOnLinux(sFilePath, sJobId, bIsJobRunning, sServerName,
				sPortNo, isPASSecure, username, password, pbsusername, pbspassword,
				datasource, userAuthToken, sPASUrl)
			// set the response on file download object for other waiting threads
			l.Log().Info("Setting downloaded file absolute path in downloadedFileObject to be used by other thread")
		}
		downloadedPbsObject.LocalFilePath = fileDownloaded
	}

	//downloadedFileObject.Unlock()
	var endTime = time.Now()
	l.Log().Info("File Download Time ", endTime.Sub(startTime))
	return fileDownloaded

}

/**
     * This method finds whether any similar request is being computed in the memory
     * or not. If it finds a similar request, it returns the same file download object
     * otherwise it builds an object for the file download request and returns it.
     *
	 **/
func GetDownloadedPASFileObject(sRemoteFilePath string, sJobId string) DownloadedPBSDSObject {
	l.Log().Info("Entering method DownloadedPBSDSObject")
	var sFilePath = utils.GetPlatformIndependentFilePath(sRemoteFilePath, false)
	l.Log().Info("Looking for existing request for file path " + sRemoteFilePath + " and job id " + sJobId)
	var downloadedPbsObject = getExistingPBSRequest(sFilePath, sJobId)
	if downloadedPbsObject.RemoteFilePath == "" {
		l.Log().Info("No existing request found, create new DownloadedFilePortDSObject")
		// No existing similar request - so build it and add it to the ongoing req list
		var downloadedPbsObjectNew = DownloadedPBSDSObject{
			RemoteFilePath:         sFilePath,
			JobId:                  sJobId,
			IOngoingRequestCounter: 1,
		}

		PBSDSDownloadConcurrentRequest.LstOngoingRequests = append(PBSDSDownloadConcurrentRequest.LstOngoingRequests, downloadedPbsObjectNew)
		return downloadedPbsObjectNew
	} else {
		l.Log().Info("Increment the counter for DownloadedFilePortDSObject")
		downloadedPbsObject.IOngoingRequestCounter = downloadedPbsObject.IOngoingRequestCounter + 1
		return downloadedPbsObject
	}
}

func getExistingPBSRequest(sRemoteFilePath string, jobId string) DownloadedPBSDSObject {
	l.Log().Info("Entering method getExistingRequest")

	for i := 0; i < len(PBSDSDownloadConcurrentRequest.LstOngoingRequests); i++ {
		if PBSDSDownloadConcurrentRequest.LstOngoingRequests[i].RemoteFilePath == sRemoteFilePath &&
			PBSDSDownloadConcurrentRequest.LstOngoingRequests[i].JobId == jobId {
			return PBSDSDownloadConcurrentRequest.LstOngoingRequests[i]
		}
	}
	l.Log().Info("Found no existing request")
	return DownloadedPBSDSObject{}
}

func downloadPBSSeriesFileOnLinux(sFilePath string, sJobId string, bIsJobRunning bool,
	sServerName string, sPortNo string, isPASSecure bool,
	username string, password string, pbsusername string, pbspassword string,
	dataSource datamodel.ResourceDataSource, userAuthToken string, sPASUrl string) string {
	l.Log().Info("Entering method downloadPBSSeriesFileOnLinux")

	var lstChangedFiles []string
	l.Log().Info("Retrieving last modified time for all series files present")
	var mapCurrentFileVsModTime = getPBSLastModificationTimeForParentDir(dataSource, sFilePath, bIsJobRunning, userAuthToken, sPASUrl)

	//TODO ad cache code

	//Download first time
	lstChangedFiles = getAllPbsFilesList(mapCurrentFileVsModTime)
	var fileDownloaded = readSeriesPBSFileFromPBSServerUserToken(sFilePath, sJobId, bIsJobRunning, sServerName, sPortNo,
		isPASSecure, username, password, pbsusername, pbspassword,
		dataSource, userAuthToken, lstChangedFiles, "", sPASUrl)
	return fileDownloaded

}

func getPBSLastModificationTimeForParentDir(dataSource datamodel.ResourceDataSource, sFilePath string, bIsJobRunning bool,
	userAuthToken string, sPASUrl string) map[string]int64 {
	l.Log().Info("Entering method getPBSLastModificationTimeForParentDir")
	var sParentDirPath = utils.GetPlatformIndependentFilePath(filepath.Dir(dataSource.FilePath), false)
	var mapFileVsModTime = make(map[string]int64)
	var pbsServer = dataSource.PbsServer
	var sJobId = ""
	l.Log().Info("Getting FileOperations port")

	l.Log().Info("User:" + pbsServer.UserName + ": Checking file access permission: " + sParentDirPath)
	var sJobStatus = pbsServer.JobStatus
	if utils.PBS_JOB_RUNNING_STATE == sJobStatus {
		sJobId = pbsServer.JobId
	}

	l.Log().Info("Calling fileList operation on job id" + sJobId + " and parent dir path " + sParentDirPath)

	var fileListResult = getPbsFileListWLM(sJobId, bIsJobRunning, sParentDirPath, userAuthToken, sPASUrl)

	if len(fileListResult) != 0 {
		l.Log().Info("Parent dir may contain other files than series files. lets filter them")
		var lstFilteredFileData = filterFileData(fileListResult, sFilePath)

		l.Log().Info("Adding last modified time in a map for all the filtered files")
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
	if bIsJobRunning && strings.Contains(sPasURL, utils.PAS_URL_VALUE) {
		sPasURL = strings.Replace(sPasURL, utils.PAS_URL_VALUE, utils.JOB_OPERATION, -1)
	} else {
		sPasURL = sPasURL + utils.REST_SERVICE_URL
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

	l.Log().Info("Downloading file from WLM using WLM API")
	if fileToWrite == "" {
		fileToWrite = getDownloadFilePath(dataSource, username, password, sFilePath)
	}

	var sUrl = buildPbsMultiDownloadUrl(bIsJobRunning, sPASUrl)
	var urlParameters = buildPbsURLParametres(filesToDownload, sFilePath, sJobId)
	var fileDownloadStartTime = time.Now()
	var zipFile = DownloadMultiFileAsZip(sUrl, urlParameters, userAuthToken, utils.GetDirPath(fileToWrite))
	var fileDownloadEndTime = time.Now()
	l.Log().Info("File Service Download Time ", fileDownloadEndTime.Sub(fileDownloadStartTime))
	var wirteTime = time.Now()
	if len(filesToDownload) != 1 {
		l.Log().Info("Downloading multifile as zip ")
		_, err := Unzip(zipFile, utils.GetDirPath(fileToWrite))
		if err != nil {
			log.Fatal(err)
		}
		e := os.Remove(zipFile)
		if e != nil {
			log.Fatal(e)
		}
	}
	l.Log().Info("File Service Download Time ", time.Since(wirteTime))
	// Get last modification time
	l.Log().Info("Getting last modification time of datasource")
	var JobState = ""
	if bIsJobRunning {
		JobState = "R"
	}
	var lastModTime = GetLastModificationTime(JobState, sJobId, sPASUrl, sFilePath, userAuthToken)
	// Set original last modification time
	l.Log().Info("Setting last modification time received from datasource on the local temp file ", fileToWrite)
	var timeInMilllis, _ = strconv.ParseInt(lastModTime, 10, 64)

	err := os.Chtimes(fileToWrite, time.Now(), time.UnixMilli(timeInMilllis))
	if err != nil {
		l.Log().Error(err)
	}
	return fileToWrite

}

func buildPbsMultiDownloadUrl(bIsJobRunning bool, sPASUrl string) string {

	if bIsJobRunning && strings.Contains(sPASUrl, utils.PAS_URL_VALUE) {
		sPASUrl = strings.Replace(sPASUrl, utils.PAS_URL_VALUE, utils.JOB_OPERATION, -1)
	} else {
		sPASUrl = sPASUrl + utils.REST_SERVICE_URL
	}
	sPASUrl = sPASUrl + "/files/downloadMulti"

	return sPASUrl

}

func buildPbsURLParametres(lstFilePath []string, sFilePath string, sJobID string) string {
	var urlParameters = ""
	var path = utils.GetDirPath(sFilePath)
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

	l.Log().Info("Entering downloadFileOnLinux downloadFile")
	//TODO Cache

	l.Log().Info("Copying files via SCP/RVP failed. Falling back to AIF file copy...")
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
	l.Log().Info("Downloading file from WLM using WLM API")
	var fileToWrite = getPbsDownloadFilePath(dataSource, username, password, sFilePath)
	l.Log().Info("sJobId:" + sJobId + "bIsJobRunning" + strconv.FormatBool(bIsJobRunning))
	var sUrl = buildDownloadUrl(sServerName, sPortNo, isPASSecure, sFilePath, sJobId, bIsJobRunning, sPASUrl)
	l.Log().Info("downloadUrl:" + sUrl)

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
	l.Log().Info("Getting last modification time of datasource")
	var JobState = ""
	if bIsJobRunning {
		JobState = "R"
	}
	var timeInMilllisString = GetLastModificationTime(JobState, sJobId, sPASUrl, sFilePath, userAuthToken)
	var timeInMilllis, _ = strconv.ParseInt(timeInMilllisString, 10, 64)

	err := os.Chtimes(fileToWrite, time.Now(), time.UnixMilli(timeInMilllis))
	if err != nil {
		l.Log().Error(err)
	}
	return fileToWrite

}

func getPbsDownloadFilePath(dataSource datamodel.ResourceDataSource, username string, passowrd string, sFilePath string) string {

	var sFilePathNew = utils.GetPlatformIndependentFilePath(sFilePath, false)

	var parentFolder = AllocateUniqueFolder(SiteConfigData.RVSConfiguration.HWE_RM_DATA_LOC+utils.RM_DOWNLOADS, "DOWNLOAD")

	var fileToWrite = AllocateFileWithGlobalPermission(utils.GetFileName(sFilePathNew), parentFolder)
	l.Log().Info("Created temp file " + fileToWrite + " to store the pbs data source")
	return fileToWrite
}

func buildDownloadUrl(sServerName string, sPortNo string, isSecure bool, serverSideFilePath string, sJobId string,
	bIsJobRunning bool, sPASUrl string) string {
	if bIsJobRunning && strings.Contains(sPASUrl, utils.PAS_URL_VALUE) {
		sPASUrl = strings.Replace(sPASUrl, utils.PAS_URL_VALUE, utils.JOB_OPERATION, -1)
	} else {
		sPASUrl = sPASUrl + utils.REST_SERVICE_URL
	}
	sPASUrl = sPASUrl + "/files/download"
	l.Log().Info("downloadUrl:" + sPASUrl)
	return sPASUrl
}

func downloadPBSFile(sUrl string, userAuthToken string) string {
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
