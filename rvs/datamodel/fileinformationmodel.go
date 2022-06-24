package datamodel

type FileInformationModel struct {
	SeriesFile       bool   `json:"seriesFile"`
	FileExtension    string `json:"fileExtension"`
	FileName         string `json:"fileName"`
	FilePath         string `json:"filePath"`
	ServerName       string `json:"serverName"`
	PasUrl           string `json:"pasUrl"`
	JobId            string `json:"jobId"`
	JobState         string `json:"jobState"`
	Id               string `json:"id"`
	LastModifiedTime string `json:"lastModifiedTime"`
	Size             string `json:"size"`
	FileUrl          string `json:"fileUrl"`
}
