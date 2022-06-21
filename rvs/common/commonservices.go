package common

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
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
	fmt.Println("pasUrl:", pasUrl)
	fmt.Println("urlstring:", urlString)
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
	fmt.Println("jobLocation", jobLocation)
	var folderExist = DoesFileExist(pasUrl, jobstate, jobId, sToken, jobLocation)
	fmt.Println("folderExist", folderExist)
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

	fmt.Println(resp.Status)
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
