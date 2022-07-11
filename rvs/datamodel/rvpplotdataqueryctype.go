package datamodel

type RVPPlotDataQueryCType struct {
	RVPSimulationQuery *RvpsimulationQuery `json:"simulationQuery"`
	RVPPlotColumnInfo  RvpPlotColumnInfo   `json:"rvpPlotColumnInfo"`
}

type RvpsimulationQuery struct {
	RVPSimulationRangeBasedQuery *RvpsimulationRangeBasedQuery `json:"simulationRangeBasedQuery"`
	RVPSimulationCountBasedQuery *RvpsimulationCountBasedQuery `json:"simulationCountBasedQuery"`
}

type RvpsimulationRangeBasedQuery struct {
	StartIndex int `json:"startIndex"`
	EndIndex   int `json:"endIndex"`
	Step       int `json:"step"`
}

type RvpsimulationCountBasedQuery struct {
	StartIndex int `json:"startIndex"`
	Count      int `json:"count"`
	Step       int `json:"step"`
}
