package graph

import "altair/rvs/datamodel"

type PlotRequestResModel struct {
	ResultFileInformationModel  resultFileInformationModel   `json:"resultFileInformationModel"`
	PlotRequestResponseModel    plotRequestResponseModel     `json:"plotRequestResponseModel"`
	LstMatchingResultFiles      []resultFileInformationModel `json:"lstMatchingResultFiles"`
	LstPlotRequestResponseModel []plotRequestResponseModel   `json:"lstPlotRequestResponseModel"`
}

type resultFileInformationModel struct {
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

type plotRequestResponseModel struct {
	PlotMetaData          plotMetaData          `json:"plotMetaData"`
	TemplateMetaDataModel templateMetaDataModel `json:"templateMetaDataModel"`
	PlotResponseModel     plotResponseModel     `json:"plotResponseModel"`
	Queries               queries               `json:"queries"`
	Responses             responses             `json:"Responses"`
	WindowPositionModel   windowPositionModel   `json:"windowPositionModel"`
}

type plotMetaData struct {
	GraphMetaData graphMetaData `json:"graphMetaData"`
	TitleMetaData titleMetaData `json:"titleMetaData"`
	ExtraMetaData extraMetaData `json:"extraMetaData"`
	UserPreferece userPreferece `json:"userPrefereces"`
}

type graphMetaData struct {
	ShowLegend     bool `json:"showLegend"`
	ShowDataPoints bool `json:"showDataPoints"`
	ShowXlogScale  bool `json:"showXlogScale"`
	ShowYlogScale  bool `json:"showYlogScale"`
}

type extraMetaData struct {
	MetaData []metaData `json:"metaData"`
}

type metaData struct {
	MetaDataName  string `json:"metaDataName"`
	MetaDataValue string `json:"metaDataValue"`
}

type userPreferece struct {
	UserPrefereces []userPrefereces `json:"userPrefereces"`
}

type userPrefereces struct {
	Name               string `json:"name"`
	CurveLineThickness string `json:"curveLineThickness"`
	CurveDatapointSize string `json:"curveDatapointSize"`
	CurveColors        string `json:"curveColors"`
	EnableDataPoint    bool   `json:"enableDataPoint"`
}

type titleMetaData struct {
	Title      string `json:"title"`
	XaxisTitle string `json:"xaxisTitle"`
	YaxisTitle string `json:"yaxisTitle"`
}

type templateMetaDataModel struct {
	TemplateId            string `json:"templateId"`
	TemplateName          string `json:"templateName"`
	FileName              string `json:"fileName"`
	FileExtension         string `json:"fileExtension"`
	ApplicationName       string `json:"applicationName"`
	UserName              string `json:"userName"`
	TemplateData          string `json:"templateData"`
	IsDefault             bool   `json:"isDefault"`
	IsSeriesFile          bool   `json:"isSeriesFile"`
	IsFilteredReqTemplate bool   `json:"isFilteredReqTemplate"`
}
type plotResponseModel struct {
	PlotAmCharts             plotAmCharts `json:"plotAmCharts"`
	TemporaryPltFilePath     string       `json:"temporaryPltFilePath"`
	NewlyAddedPltBlocksCount int          `json:"newlyAddedPltBlocksCount"`
}

type plotAmCharts struct {
	ExportPlotDataUrl         string        `json:"exportPlotDataUrl"`
	ChartHtmlUrl              string        `json:"chartHtmlUrl"`
	ExportPlotDataRelativeUrl string        `json:"exportPlotDataRelativeUrl"`
	ChartHtmlRelativeUrl      string        `json:"chartHtmlRelativeUrl"`
	PlotFileRelativePath      string        `json:"plotFileRelativePath"`
	PlotDataModel             plotDataModel `json:"plotDataModel"`
}

type plotDataModel struct {
	XaxisNegative       bool     `json:"xaxisNegative"`
	YaxisNegative       bool     `json:"yaxisNegative"`
	CurveNames          []string `json:"curveNames"`
	LegendNames         []string `json:"legendNames"`
	NumberOfCurvePoints int      `json:"numberOfCurvePoints"`
	DataPoints          string   `json:"dataPoints"`
	LogXdataPoints      string   `json:"logXdataPoints"`
	LogYdataPoints      string   `json:"exlogYdataPointsportPlotDataUrl"`
	LogXlogYdataPoints  string   `json:"logXlogYdataPoints"`
}

type queries struct {
	ResultDataSource []datamodel.ResourceDataSource `json:"resultDataSource"`
	Query            []Query                        `json:"query"`
}

type Query struct {
	ResultDataSourceRef []resultDataSourceRef `json:"resultDataSourceRef"`
	OutputSource        outputSource          `json:"outputSource"`

	PlotResultQuery      plotResultQuery      `json:"plotResultQuery"`
	RvpPlotDataQuery     rvpPlotDataQuery     `json:"rvpPlotDataQuery"`
	AnimationResultQuery animationResultQuery `json:"animationResultQuery"`
	//Custom               custom               `json:"custom"`
	Session           session `json:"session"`
	IsCachingRequired bool    `json:"isCachingRequired"`
	Id                int     `json:"id"`
	VarName           string  `json:"varName"`
}

type plotResultQuery struct {
	DataQuery       dataQuery       `json:"dataQuery"`
	ExpressionQuery expressionQuery `json:"expressionQuery"`
}

type dataQuery struct {
	IsRawDataRequired bool             `json:"isRawDataRequired"`
	SimulationQuery   simulationQuery  `json:"simulationQuery"`
	SimulationFilter  simulationFilter `json:"simulationFilter"`
	StrcQuery         strcQuery        `json:"strcQuery"`
	InlineQuery       inlineQuery      `json:"inlineQuery"`
}

type expressionQuery struct {
	ScriptDataSource scriptDataSource `json:"scriptDataSource"`
}
type scriptDataSource struct {
	Id                   string `json:"id"`
	Label                string `json:"label"`
	IsForceRefresh       bool   `json:"isForceRefresh"`
	LastModificationTime int64  `json:"lastModificationTime"`
	SeriesFile           int64  `json:"seriesFile"`
}
type simulationQuery struct {
	SimulationRangeBasedQuery simulationRangeBasedQuery `json:"simulationRangeBasedQuery"`
	SimulationCountBasedQuery simulationCountBasedQuery `json:"simulationCountBasedQuery"`
}

type simulationRangeBasedQuery struct {
	StartIndex int `json:"startIndex"`
	EndIndex   int `json:"endIndex"`
	Step       int `json:"step"`
}

type simulationCountBasedQuery struct {
	StartIndex int `json:"startIndex"`
	Count      int `json:"count"`
	Step       int `json:"step"`
}

type simulationFilter struct {
	Start int   `json:"start"`
	End   int64 `json:"end"`
	Step  int64 `json:"step"`
}

type strcQuery struct {
	DistantRequest    distantRequest    `json:"distantRequest"`
	ContiguousRequest contiguousRequest `json:"contiguousRequest"`
	Subcase           subcase           `json:"subcase"`
	Type              Type              `json:"type"`
	//Statistic         statistic         `json:"statistic"`
	//Sampling          sampling          `json:"sampling"`
}

type inlineQuery struct {
	Title      string `json:"title"`
	Expression string `json:"enexpressiond"`
	InlineData string `json:"inlineData"`
}

type distantRequest struct {
	Component   component   `json:"component"`
	DataRequest dataRequest `json:"dataRequest"`
}

type component struct {
	Name string `json:"name"`
}

type dataRequest struct {
	Name string `json:"name"`
}

type contiguousRequest struct {
	DataRequestIndex dataRequestIndex `json:"dataRequestIndex"`
	ComponentIndex   componentIndex   `json:"componentIndex"`
	TimeStep         timeStep         `json:"timeStep"`
}

type dataRequestIndex struct {
	Start int `json:"start"`
	End   int `json:"end"`
}
type componentIndex struct {
	Start int `json:"start"`
	End   int `json:"end"`
}
type timeStep struct {
	Index int `json:"index"`
}

type subcase struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type Type struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type windowPositionModel struct {
	ColumnNumber       int  `json:"columnNumber"`
	PageNumber         int  `json:"pageNumber"`
	RowNumber          int  `json:"rowNumber"`
	HorizontalExpanded bool `json:"horizontalExpanded"`
	VerticalExpanded   bool `json:"verticalExpanded"`
	FullScreenExpanded bool `json:"fullScreenExpanded"`
	IsCommonBlock      bool `json:"isCommonBlock"`
}

type rvpPlotDataQuery struct {
	SimulationQuery   simulationQuery   `json:"simulationQuery"`
	RvpPlotColumnInfo rvpPlotColumnInfo `json:"rvpPlotColumnInfo"`
}

type rvpPlotColumnInfo struct {
	PlotName    string   `json:"plotName"`
	ColumnNames []string `json:"columnNames"`
	ColumnName  string   `json:"columnName"`
}

type animationResultQuery struct {
	H3DQuery h3DQuery `json:"h3DQuery"`
}

type session struct {
	Id            string `json:"id"`
	RetainSession bool   `json:"retainSession"`
}
type h3DQuery struct {
	ModelDataSource  datamodel.ResourceDataSource `json:"modelDataSource"`
	ConfigDataSource datamodel.ResourceDataSource `json:"configDataSource"`
	//InlineH3DQuery        inlineH3DQuery   `json:"inlineH3DQuery"`
	TranslateAll          bool    `json:"translateAll"`
	CompressionPercentage float32 `json:"compressionPercentage"`
	SaveOnlyModel         bool    `json:"saveOnlyModel"`
	SaveOnlyResult        bool    `json:"saveOnlyResult"`
	IncludeModel          bool    `json:"includeModel"`
}

type resultDataSourceRef struct {
	Id             string `json:"id"`
	IdUsedInScript int    `json:"idUsedInScript"`
}

type outputSource struct {
	Path            string `json:"path"`
	FileName        string `json:"fileName"`
	CreateNewFolder bool   `json:"createNewFolder"`
	OverwriteFile   bool   `json:"overwriteFile"`
}

type Res struct {
	Responses responses `json:"Responses"`
}
type responses struct {
	Responselist []response `json:"Response"`
}

type response struct {
	ResponseData responseData `json:"ResponseData"`
	Id           string       `json:"id"`
}

type responseData struct {
	DataSource dataSource `json:"DataSource"`
	Tag        tag        `json:"tag"`
}

type dataSource struct {
	Type        string    `json:"type"`
	Items       []float64 `json:"items"`
	NoOfItems   int       `json:"noOfItems"`
	IsRowVector bool      `json:"isRowVector"`
}
type tag struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	//dataType dataType    `json:"dataType"`
}

type PlotTemporaryModel struct {
	PlotMetaData   plotMetaData `json:"plotMetaData"`
	lstCurveNames  []string
	lstCurvesData  [][]float64
	lstLegendNames []string
}

type plotinstance struct {
	Instances Instances `json:"Instances"`
}
type Instances struct {
	PLT []PLT `json:"PLT"`
}
type PLT struct {
	PlotMetaData          plotMetaData          `json:"plotMetaData"`
	TemplateMetaDataModel templateMetaDataModel `json:"templateMetaDataModel"`
	PlotResponseModel     plotResponseModel     `json:"plotResponseModel"`
	Queries               queries               `json:"queries"`
	Responses             responses             `json:"Responses"`
	WindowPositionModel   windowPositionModel   `json:"windowPositionModel"`
}

type MatchingFiles struct {
	Success  bool   `json:"success"`
	Data     data   `json:"data"`
	ExitCode string `json:"exitCode"`
}

type data struct {
	Files []files `json:"files"`
}

type files struct {
	Created     int64  `json:"created"`
	Filename    string `json:"filename"`
	FileExt     string `json:"fileExt"`
	Modified    int64  `json:"modified"`
	Owner       string `json:"owner"`
	Size        int64  `json:"size"`
	Type        string `json:"type"`
	IsWritable  bool   `json:"isWritable"`
	AbsPath     string `json:"absPath"`
	HasChildren bool   `json:"hasChildren"`
	IsReadable  bool   `json:"isReadable"`
	IisHidden   bool   `json:"isHidden"`
}
