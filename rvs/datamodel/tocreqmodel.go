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

type subcase struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Id    string `json:"id"`
}

type Type struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Id    int    `json:"id"`
}
type filter struct {
	Id      int    `json:"id"`
	GetNext bool   `json:"getNext"`
	Start   string `json:"start"`
	Count   int    `json:"count"`
}
