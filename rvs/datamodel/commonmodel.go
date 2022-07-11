package datamodel

type subcase struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Id    string `json:"id"`
}

type Type struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
	Id    *int   `json:"id"`
}

type rvpPlotColumnInfo struct {
	PlotName    string   `json:"plotName"`
	ColumnNames []string `json:"columnNames"`
	ColumnName  string   `json:"columnName"`
}
