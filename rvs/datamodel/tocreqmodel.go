package datamodel

type TOCRequest struct {
	User               string             `json:"user"`
	Pwd                string             `json:"pwd"`
	PostProcessingType string             `json:"postProcessingType"`
	ResultDataSource   ResourceDataSource `json:"resultDataSource"`
	ModelDataSource    string             `json:"modelDataSource"`
	Custom             string             `json:"custom"`
	IsCachingRequired  string             `json:"isCachingRequired"`
	SchemaVersion      string             `json:"schemaVersion"`
	PlotFilter         plotFilter         `json:"plotFilter"`
}

type plotFilter struct {
	Subcase subcase `json:"subcase"`
	Type    Type    `json:"type"`
	Filter  filter  `json:"filter"`
}

type filter struct {
	Id      *int   `json:"id"`
	GetNext bool   `json:"getNext"`
	Start   string `json:"start"`
	Count   int64  `json:"count"`
}
