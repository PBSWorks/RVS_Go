package datamodel

import "encoding/xml"

type Plugin struct {
	XMLName      xml.Name     `xml:"plugin"`
	DataProvider DataProvider `xml:"DataProvider"`
}

type DataProvider struct {
	XMLName        xml.Name       `xml:"DataProvider"`
	SupportedFiles SupportedFiles `xml:"SupportedFiles"`
}

type SupportedFiles struct {
	XMLName xml.Name `xml:"SupportedFiles"`
	File    []File   `xml:"File"`
}

type File struct {
	IsDefault                   bool              `xml:"isDefault,attr"`
	SupportsDirectPlotOperation bool              `xml:"supportsDirectPlotOperation,attr"`
	Value                       string            `xml:",chardata"`
	Pattern                     string            `xml:"Pattern"`
	ParsingStrategies           ParsingStrategies `xml:"ParsingStrategies"`
	Translator                  Translator        `xml:"Translator"`
}

type ParsingStrategies struct {
	XMLName         xml.Name          `xml:"ParsingStrategies"`
	ParsingStrategy []ParsingStrategy `xml:"ParsingStrategy"`
}
type ParsingStrategy struct {
	XMLName     xml.Name    `xml:"ParsingStrategy"`
	Id          string      `xml:"Id,attr"`
	ColumnNames ColumnNames `xml:"ColumnNames"`
	DataPoints  DataPoints  `xml:"DataPoints"`
	Comments    Comments    `xml:"Comments"`
}

type ColumnNames struct {
	XMLName          xml.Name `xml:"ColumnNames"`
	Prefix           string   `xml:"Prefix"`
	Delimiter        string   `xml:"Delimiter"`
	ColumnNamePrefix string   `xml:"ColumnNamePrefix"`
}
type DataPoints struct {
	XMLName   xml.Name `xml:"DataPoints"`
	Prefix    string   `xml:"Prefix"`
	Delimiter string   `xml:"Delimiter"`
	Locale    Locale   `xml:"Locale"`
}

type Locale struct {
	XMLName  xml.Name `xml:"Locale"`
	Language string   `xml:"Language"`
	Country  string   `xml:"Country"`
}

type Comments struct {
	XMLName xml.Name `xml:"Comments"`
	Prefix  string   `xml:"Prefix"`
}

type Translator struct {
	XMLName                          xml.Name         `xml:"Translator"`
	ScriptAbsolutePath               string           `xml:"ScriptAbsolutePath"`
	TemporaryOutputFileExtension     string           `xml:"TemporaryOutputFileExtension"`
	ResultFileAbsolutePathArgName    string           `xml:"ResultFileAbsolutePathArgName"`
	TemporaryFileAbsolutePathArgName string           `xml:"TemporaryFileAbsolutePathArgName"`
	ScriptParameters                 ScriptParameters `xml:"ScriptParameters"`
}

type ScriptParameters struct {
	XMLName         xml.Name          `xml:"ScriptParameters"`
	ScriptParameter []ScriptParameter `xml:"ScriptParameter"`
}
type ScriptParameter struct {
	XMLName xml.Name `xml:"ScriptParameter"`
	Key     string   `xml:"key"`
	Value   string   `xml:"value"`
}

type SupportedFilePatterns struct {
	Pattern []Pattern `json:"pattern"`
}
type Pattern struct {
	Value                     string `json:"value"`
	PpType                    string `json:"ppType"`
	DefaultPostProcessingType string `json:"defaultPostProcessingType"`
}

type SupportedFilePatternOutput struct {
	MapFilePatterns MapFilePatterns `json:"mapFilePatterns"`
}
type MapFilePatterns struct {
	WLMSERVER WLMSERVER `json:"WLMSERVER"`
}
type WLMSERVER struct {
	IsHWConfigured bool      `json:"isHWConfigured"`
	Pattern        []Pattern `json:"pattern"`
}
