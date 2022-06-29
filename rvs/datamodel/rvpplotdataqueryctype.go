package datamodel

type RVPPlotDataQueryCType struct {
	RVPSimulationQuery *RvpsimulationQuery `json:"simulationQuery"`
	RVPPlotColumnInfo  rvpPlotColumnInfo   `json:"rvpPlotColumnInfo"`
}

type RvpsimulationQuery struct {
	RVPSimulationRangeBasedQuery *rvpsimulationRangeBasedQuery `json:"simulationRangeBasedQuery"`
	RVPSimulationCountBasedQuery *rvpsimulationCountBasedQuery `json:"simulationCountBasedQuery"`
}

type rvpsimulationRangeBasedQuery struct {
	StartIndex int `json:"startIndex"`
	EndIndex   int `json:"endIndex"`
	Step       int `json:"step"`
}

type rvpsimulationCountBasedQuery struct {
	StartIndex int `json:"startIndex"`
	Count      int `json:"count"`
	Step       int `json:"step"`
}

type rvpPlotColumnInfo struct {
	PlotName    string   `json:"plotName"`
	ColumnNames []string `json:"columnNames"`
	ColumnName  string   `json:"columnName"`
}
