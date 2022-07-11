package datamodel

type InstanceQueryModel struct {
	ResultFileInformationModel ResultFileInformationModel `json:"fileInfoModel"`
	WindowInfoModel            windowInfoModel            `json:"windowInfoModel"`
}

type windowInfoModel struct {
	WindowPositionModel      WindowPositionModel      `json:"windowPosition"`
	WindowView               string                   `json:"windowView"`
	PlotRequestResponseModel PlotRequestResponseModel `json:"plotRequestResponseModel"`
}
