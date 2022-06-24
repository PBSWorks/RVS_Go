package datamodel

type SupportedRVPFilesModel struct {
	ListRVPFileModel []RVPFileModel `json:"listRVPFileModel"`
}

type RVPFileModel struct {
	IsDefault                   bool                       `json:"isDefault"`
	SupportsDirectPlotOperation bool                       `json:"supportsDirectPlotOperation"`
	Pattern                     string                     `json:"pattern"`
	RvpFileTranslator           RvpFileTranslator          `json:"rvpFileTranslator"`
	FileParsingStrategiesModel  FileParsingStrategiesModel `json:"fileParsingStrategiesModel"`
}

type RvpFileTranslator struct {
	ScriptAbsolutePath                  string              `json:"scriptAbsolutePath"`
	ResultFileAbsolutePathArgName       string              `json:"resultFileAbsolutePathArgName"`
	TemporaryRVPFileAbsolutePathArgName string              `json:"temporaryRVPFileAbsolutePathArgName"`
	TemporaryOutputRVPFileExtension     string              `json:"temporaryOutputRVPFileExtension"`
	ScriptParameters                    (map[string]string) `json:"scriptParameters"`
}

type FileParsingStrategiesModel struct {
	ListFileParsingStrategyModel []FileParsingStrategyModel `json:"listFileParsingStrategyModel"`
}

type FileParsingStrategyModel struct {
	Id                     string                 `xml:"Id,attr"`
	ColumnNamesParserModel ColumnNamesParserModel `json:"columnNamesParserModel"`
	CommentsParserModel    CommentsParserModel    `json:"commentsParserModel"`
	DataPointsParserModel  DataPointsParserModel  `json:"dataPointsParserModel"`
}

type ColumnNamesParserModel struct {
	Prefix           string `json:"prefix"`
	Delimiter        string `json:"delimiter"`
	ColumnNamePrefix string `json:"columnNamePrefix"`
}

type CommentsParserModel struct {
	Prefix string `json:"prefix"`
}

type DataPointsParserModel struct {
	Prefix       string       `json:"prefix"`
	Delimiter    string       `json:"delimiter"`
	NumberLocale NumberLocale `json:"numberLocale"`
}

type NumberLocale struct {
	Language string `json:"Language"`
	Country  string `json:"Country"`
}
