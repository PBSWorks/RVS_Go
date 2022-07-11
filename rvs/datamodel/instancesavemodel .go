package datamodel

type InstanceSaveModel struct {
	PlotSaveModelList          []plotSaveModelList        `json:"plotSaveModelList"`
	ResultFileInformationModel ResultFileInformationModel `json:"fileInformationModel"`
	CanOverwrite               bool                       `json:"canOverwrite"`
}

type plotSaveModelList struct {
	PlotMetaData        PlotMetaData        `json:"plotMetaData"`
	PlotResponseModel   PlotResponseModel   `json:"plotResponseModel"`
	WindowPositionModel WindowPositionModel `json:"windowPositionModel"`
}
