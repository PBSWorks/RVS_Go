package main

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"altair/rvs/graph"
	"altair/rvs/toc"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var tocRequest datamodel.TOCRequest

func getToc(sServerName string, resultfilepath string, sIsSeriesFile string,
	sJobId string, sJobState string, token string, pasURL string) (string, error) {

	var tocType = tocRequest.PostProcessingType
	if tocType == "PLOT" {
		return toc.GetPlotToc(sServerName, resultfilepath, sIsSeriesFile,
			tocRequest, sJobId, sJobState, token, pasURL, "", "")
	} else if tocType == "ANIMATION" {
		return toc.GetAnimationToc(sServerName, resultfilepath, sIsSeriesFile,
			tocRequest, sJobId, sJobState, token, pasURL)
	} else {
		return "", nil
	}
}

func getFilterToc(sServerName string, resultfilepath string, sIsSeriesFile string,
	tocRequest datamodel.TOCRequest, sJobId string, sJobState string, token string, pasURL string) (string, error) {

	return toc.GetPlotToc(sServerName, resultfilepath, sIsSeriesFile,
		tocRequest, sJobId, sJobState, token, pasURL, tocRequest.PlotFilter.Subcase.Name, tocRequest.PlotFilter.Type.Name)
}

func getModelToc(resultFilepath string, jobid string, jobstate string, server string, pasURL string,
	token string, username string, password string) (string, error) {

	return toc.GetModelToc(resultFilepath, jobid, jobstate, server, pasURL, token, username, password)
}

func getRVPToc(fileInformationModel datamodel.FileInformationModel, sToken string, username string, password string) (string, error) {

	return toc.GetRVPToc(fileInformationModel, sToken, username, password)
}

func getPlotGraph(plotRequestResModel graph.PlotRequestResModel, plotRequestCaller string, username string, password string, token string) string {

	return graph.GetPlotGraphExtractor(plotRequestResModel, plotRequestCaller, username, password, token)
}

func getSupportedFilePatternsForAllServers(token string) string {
	return common.GetSupportedFilePatternsForAllServers(token)
}
func getSeriesFilePatternsForAllServer(token string) string {
	return common.GetSupportedSeriesFilePatterns(token)
}

func GetHWConfigDetails() string {
	return common.GetHWComposeConfigDetails()
}

func viewPlotData(requestData []byte, pasURL string, sToken string, username string) string {
	return graph.ViewPLT(requestData, pasURL, sToken, username)
}

func saveInstanceData(requestData []byte, pasURL string, sToken string) (string, error) {
	return graph.SaveInstance(requestData, pasURL, sToken)
}
func refreshPlotData(sRequestData []byte, username string, password string, sToken string) string {
	return graph.RefreshPlt(sRequestData, username, password, sToken)
}
func overlayPlotData(sRequestData []byte, username string, password string, sToken string) (string, error) {
	return graph.OverlayPlt(sRequestData, username, password, sToken)
}

func getTOCData(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &tocRequest)
	query := r.URL.Query()
	resultfilepath := query.Get("resultfilepath")
	sJobId := query.Get("jobid")
	sJobState := query.Get("jobstate")
	sServerName := query.Get("server")
	sIsSeriesFile := query.Get("seriesfile")
	pasURL := query.Get("pasURL")
	token := r.Header.Get("Authorization")
	var output, err = getToc(sServerName, resultfilepath, sIsSeriesFile,
		sJobId, sJobState, strings.TrimSpace(token), pasURL)

	if err != nil {
		var tocErr *exception.RVSError
		switch {
		case errors.As(err, &tocErr):
			w.Header().Set("error-code", tocErr.Errorcode)
			w.Header().Set("error-type", tocErr.Errortype)
			w.Header().Set("error-details", tocErr.Errordetails)
		default:
			log.Printf("unexpected overlay plot error: %s\n", err)
		}

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func getTOCFilterData(w http.ResponseWriter, r *http.Request) {

	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &tocRequest)

	query := r.URL.Query()

	resultfilepath := query.Get("resultfilepath")
	sJobId := query.Get("jobid")
	sJobState := query.Get("jobstate")
	sServerName := query.Get("server")
	sIsSeriesFile := query.Get("seriesfile")
	pasURL := query.Get("pasURL")
	token := query.Get("Authorization")
	var output, err = getFilterToc(sServerName, resultfilepath, sIsSeriesFile,
		tocRequest, sJobId, sJobState, strings.TrimSpace(token), pasURL)
	if err != nil {
		var filtertocErr *exception.RVSError
		switch {
		case errors.As(err, &filtertocErr):
			w.Header().Set("error-code", filtertocErr.Errorcode)
			w.Header().Set("error-type", filtertocErr.Errortype)
			w.Header().Set("error-details", filtertocErr.Errordetails)
		default:
			log.Printf("unexpected overlay plot error: %s\n", err)
		}

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

func getModelTOCData(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &tocRequest)

	query := r.URL.Query()
	jobid := query.Get("jobid")
	jobstate := query.Get("jobstate")
	server := query.Get("server")
	modelfilepath := query.Get("modelfilepath")
	pasURL := query.Get("pasURL")
	token := r.Header.Get("Authorization")
	var output, err = getModelToc(modelfilepath, jobid, jobstate, server, pasURL, strings.TrimSpace(token), tocRequest.User, tocRequest.Pwd)

	if err != nil {
		var tocErr *exception.RVSError
		switch {
		case errors.As(err, &tocErr):
			w.Header().Set("error-code", tocErr.Errorcode)
			w.Header().Set("error-type", tocErr.Errortype)
			w.Header().Set("error-details", tocErr.Errordetails)
		default:
			log.Printf("unexpected overlay plot error: %s\n", err)
		}

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func getRVPTOCData(w http.ResponseWriter, r *http.Request) {
	var fileInformationModel datamodel.FileInformationModel
	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &fileInformationModel)
	token := r.Header.Get("Authorization")
	var output, err = getRVPToc(fileInformationModel, strings.TrimSpace(token), "pbsworks", "admin@123")

	if err != nil {
		var tocErr *exception.RVSError
		switch {
		case errors.As(err, &tocErr):
			w.Header().Set("error-code", tocErr.Errorcode)
			w.Header().Set("error-type", tocErr.Errortype)
			w.Header().Set("error-details", tocErr.Errordetails)
		default:
			log.Printf("unexpected overlay plot error: %s\n", err)
		}

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func getPlotGraphData(w http.ResponseWriter, r *http.Request) {
	var plotRequestResModel graph.PlotRequestResModel

	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &plotRequestResModel)

	query := r.URL.Query()
	token := r.Header.Get("Authorization")
	plotRequestCaller := query.Get("plotRequestCaller")
	var output = getPlotGraph(plotRequestResModel, plotRequestCaller, "pbsworks", "admin@123", token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func getSupportedFilePatternsForAllServersNew(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Authorization")
	var output = getSupportedFilePatternsForAllServers(token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func isHyperWorksComposeConfigured(w http.ResponseWriter, r *http.Request) {
	var output = GetHWConfigDetails()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func getSeriesFilePatternsForAllServerNew(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Authorization")
	var output = getSeriesFilePatternsForAllServer(token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}
func saveInstance(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	query := r.URL.Query()
	token := r.Header.Get("Authorization")
	pasURL := query.Get("pasURL")
	var output, err = saveInstanceData(reqBody, pasURL, token)
	if err != nil {
		var saveplotErr *exception.RVSError
		switch {
		case errors.As(err, &saveplotErr):
			w.Header().Set("error-code", saveplotErr.Errorcode)
			w.Header().Set("error-type", saveplotErr.Errortype)
			w.Header().Set("error-details", saveplotErr.Errordetails)
		default:
			log.Printf("unexpected save plot error: %s\n", err)
		}

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func viewPlot(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	query := r.URL.Query()
	token := r.Header.Get("Authorization")
	pasURL := query.Get("pasURL")
	var output = viewPlotData(reqBody, pasURL, token, "pbsworks")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func refreshPlot(w http.ResponseWriter, r *http.Request) {

	reqBody, _ := ioutil.ReadAll(r.Body)
	token := r.Header.Get("Authorization")
	var output = refreshPlotData(reqBody, "pbsworks", "admin@123", token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func overlayPlot(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	token := r.Header.Get("Authorization")

	var output, err = overlayPlotData(reqBody, "pbsworks", "admin@123", token)
	if err != nil {
		var saveplotErr *exception.RVSError
		switch {
		case errors.As(err, &saveplotErr):
			w.Header().Set("error-code", saveplotErr.Errorcode)
			w.Header().Set("error-type", saveplotErr.Errortype)
			w.Header().Set("error-details", saveplotErr.Errordetails)
		default:
			log.Printf("unexpected overlay plot error: %s\n", err)
		}

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

func main() {
	common.Readconfigfile()
	r := mux.NewRouter()
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/toc/result", getTOCData).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/toc/result/filter", getTOCFilterData).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/toc/model", getModelTOCData).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/rvp/toc", getRVPTOCData).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/plot/data", getPlotGraphData).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/allServersFilePatternsNew", getSupportedFilePatternsForAllServersNew).Methods("GET")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/getHWConfigDetails", isHyperWorksComposeConfigured).Methods("GET")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/allserversseriesfilepatternsNew", getSeriesFilePatternsForAllServerNew).Methods("GET")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/save/instance", saveInstance).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/rvp/plt/view", viewPlot).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/refresh/plot", refreshPlot).Methods("POST")
	r.HandleFunc("/pbsworks/api/resultmanagerservice/rest/rmservice/plot/overlay", overlayPlot).Methods("POST")
	log.Fatal(http.ListenAndServe(":8083", r))
}
