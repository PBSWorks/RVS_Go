package common

import (
	"altair/rvs/utils"
	"encoding/xml"
	"io/ioutil"
	"os"

	l "altair/rvs/globlog"
)

var SiteConfigData SiteConfig

type SiteConfig struct {
	XMLName           xml.Name          `xml:"SiteConfig"`
	SiteConfig        []Products        `xml:"Products"`
	RMServers         []PAServerURL     `xml:"RMServers"`
	Environment       string            `xml:"Environment"`
	RVSConfiguration  RVSConfiguration  `xml:"RVSConfiguration"`
	SeriesResultFiles SeriesResultFiles `xml:"SeriesResultFiles"`
	Cache             Cache             `xml:"Cache"`
}

type Products struct {
	XMLName  xml.Name  `xml:"Products"`
	Products []Product `xml:"Product"`
}

type Product struct {
	XMLName        xml.Name `xml:"Product"`
	Id             string   `xml:"id,attr"`
	DefaultVersion string   `xml:"defaultVersion,attr"`
	Version        Version  `xml:"Version"`
}

type PAServerURL struct {
	PAServerURL string `xml:"PAServerURL"`
}

type Version struct {
	XMLName  xml.Name `xml:"Version"`
	Id       string   `xml:"id,attr"`
	Location string   `xml:"location,attr"`
}

type RVSConfiguration struct {
	HWE_RM_DATA_LOC string `xml:"HWE_RM_DATA_LOC"`
}

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

type Cache struct {
	XMLName    xml.Name   `xml:"Cache"`
	Enabled    string     `xml:"enabled,attr"`
	Properties Properties `xml:"Properties"`
}

type Properties struct {
	XMLName  xml.Name   `xml:"Properties"`
	Property []Property `xml:"Property"`
}

type Property struct {
	XMLName xml.Name `xml:"Property"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

func Readconfigfile() {

	// Open our xmlFile
	xmlFile, err := os.Open(utils.GetRSHome() + "/config" + utils.Siteconfigfile)
	// // if we os.Open returns an error then handle it
	if err != nil {
		l.Log().Error(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)
	// we initialize our Users array

	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &SiteConfigData)

}

func GetProductInstallationLocation(productId string) string {
	var location string = ""
	for i := 0; i < len(SiteConfigData.SiteConfig[0].Products); i++ {
		if SiteConfigData.SiteConfig[0].Products[i].Id == productId {
			location = SiteConfigData.SiteConfig[0].Products[i].Version.Location
		}
	}
	return location
}

func GetDBUrl() string {
	for i := 0; i < len(SiteConfigData.Cache.Properties.Property); i++ {
		if SiteConfigData.Cache.Properties.Property[i].Name == "gorm.datasource.url" {
			return SiteConfigData.Cache.Properties.Property[i].Value
		}
	}
	return ""
}
