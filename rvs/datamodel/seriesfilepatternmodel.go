package datamodel

import "encoding/xml"

type SeriesResultFiles struct {
	XMLName    xml.Name     `xml:"SeriesResultFiles"`
	ResultFile []ResultFile `xml:"ResultFile"`
}

type ResultFile struct {
	XMLName               xml.Name `xml:"ResultFile"`
	SeriesPattern         string   `xml:"seriesPattern,attr"`
	BasenamePattern       string   `xml:"basenamePattern,attr"`
	SeriesWildcardPattern string   `xml:"seriesWildcardPattern,attr"`
}

type SeriespatternFile struct {
	ListSupportedRvsSeriesFilePattern []SeriesPattern `json:"listSupportedRvsSeriesFilePattern"`
}
type SeriesPattern struct {
	SeriesPattern string `json:"seriesPattern"`
}
