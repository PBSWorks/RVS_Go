package datamodel

type ResourceDataSource struct {
	Custom               custom         `json:"custom"`
	Id                   string         `json:"id"`
	LastModificationTime string         `json:"lastModificationTime"`
	SeriesFile           bool           `json:"seriesFile"`
	FileName             string         `json:"fileName"`
	FilePath             string         `json:"filePath"`
	FileSize             string         `json:"fileSize"`
	FileIdentifier       string         `json:"fileIdentifier"`
	PbsServer            PbsServer      `json:"pbsServer"`
	FilePortServer       FilePortServer `json:"FilePortServer"`
}

type custom struct {
	Any []string `json:"any"`
}

type PbsServer struct {
	SessionId          string `json:"sessionId"`
	BaseUrl            string `json:"sBaseUrl"`
	PasURL             string `json:"pasURL"`
	Server             string `json:"server"`
	Port               string `json:"port"`
	JobStatus          string `json:"jobStatus"`
	Vnode              string `json:"vnode"`
	JobId              string `json:"jobId"`
	UserName           string `json:"userName"`
	UserPassword       string `json:"userPassword"`
	DirectlyAccessible bool   `json:"directlyAccessible"`
	IsSecure           bool   `json:"isSecure"`
	AuthorizationToken string `json:"authorizationToken"`
}

type FilePortServer struct {
	Name               string `json:"name"`
	UserName           string `json:"userName"`
	UserPassword       string `json:"userPassword"`
	IsSecure           bool   `json:"isSecure"`
	Port               string `json:"port"`
	AuthorizationToken string `json:"authorizationToken"`
	PasUrl             string `json:"pasUrl"`
}
