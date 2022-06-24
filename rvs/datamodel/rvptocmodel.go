package datamodel

type TOCForResultCType struct {
	RVPToc RVPToc `json:"rvpToc"`
	//SupportedPPType PostProcessingTypeSType `json:"supportedPPType"`
}

type RVPToc struct {
	RVPPlots []RVPPlotCType `json:"rvpPlots"`
}
type RVPPlotCType struct {
	RvpPlotColumnInfo RvpPlotColumnInfo `json:"rvpPlotColumnInfo"`
	Simulations       Simulations       `json:"ussimulationser"`
}

type RvpPlotColumnInfo struct {
	PlotName    string   `json:"plotName"`
	ColumnNames []string `json:"columnNames"`
	ColumnName  string   `json:"columnName"`
}

type Simulations struct {
	Delta      int64 `json:"delta"`
	Count      int   `json:"count"`
	StartIndex int   `json:"startIndex"`
	StartVal   int64 `json:"startVal"`
}
