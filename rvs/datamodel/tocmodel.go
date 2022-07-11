package datamodel

type Plots struct {
	Plot            *Plot   `json:"Plot"`
	Animation       *string `json:"Animation"`
	Model           *string `json:"Model"`
	RvpToc          *RVPToc `json:"rvpToc"`
	SupportedPPType string  `json:"SupportedPPType"`
	Custom          *string `json:"Custom"`
}

type Plot struct {
	Subcase []Subcase `json:"Subcase"`
}

type Subcase struct {
	Index       int            `json:"index"`
	Name        string         `json:"name"`
	Simulations TOCSimulations `json:"Simulations"`
	Type        []TOCType      `json:"Type"`
}
type TOCSimulations struct {
	Count      int `json:"count"`
	IndexStart int `json:"indexStart"`
}

type TOCType struct {
	Index             int              `json:"index"`
	Name              string           `json:"name"`
	IsRequestFiltered bool             `json:"isRequestFiltered"`
	Request           []Request        `json:"request"`
	RequestsOverview  RequestsOverview `json:"RequestsOverview"`
	Component         []Component      `json:"Component"`
}

type Request struct {
	ReqType    string `json:"type"`
	IndexStart int64  `json:"indexStart"`
	NameStart  string `json:"nameStart"`
	NoOfPoints int64  `json:"noOfPoints"`
	Index      int64  `json:"index"`
	Name       string `json:"name"`
}
type RequestsOverview struct {
	StartReqName string `json:"startReqName"`
	EndReqName   string `json:"endReqName"`
	NoOfRequests int    `json:"noOfRequests"`
}

type Component struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}
