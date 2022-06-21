package graph

type InstanceQueryModel struct {
	ResultFileInformationModel resultFileInformationModel `json:"fileInfoModel"`
	WindowInfoModel            windowInfoModel            `json:"windowInfoModel"`
}

type windowInfoModel struct {
	WindowPositionModel      windowPositionModel      `json:"windowPosition"`
	WindowView               string                   `json:"windowView"`
	PlotRequestResponseModel plotRequestResponseModel `json:"plotRequestResponseModel"`
}
