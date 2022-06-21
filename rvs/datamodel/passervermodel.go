package datamodel

type PASServerJobModel struct {
	JobId       string `json:"jobId"`
	JobState    string `json:"jobState"`
	ServerName  string `json:"serverName"`
	JobLocation string `json:"jobLocation"`
	PasURL      string `json:"pasURL"`
}
