package datamodel

type Templates struct {
	Template Template `json:"Templates"`
}
type Template struct {
	TPLT TPLT `json:"TPLT "`
}
type TPLT struct {
	PlotMetaData PlotMetaData `json:"plotMetaData"`
	Queries      Queries      `json:"queries"`
}
