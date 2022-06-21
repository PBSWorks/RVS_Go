package graph

type InstanceSaveModel struct {
	PlotSaveModelList          []plotSaveModelList        `json:"plotSaveModelList"`
	ResultFileInformationModel resultFileInformationModel `json:"fileInformationModel"`
	CanOverwrite               bool                       `json:"canOverwrite"`
}

type plotSaveModelList struct {
	PlotMetaData        plotMetaData        `json:"plotMetaData"`
	PlotResponseModel   plotResponseModel   `json:"plotResponseModel"`
	WindowPositionModel windowPositionModel `json:"windowPositionModel"`
}
