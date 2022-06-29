package common

import (
	"altair/rvs/datamodel"
	"regexp"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func IsCommentLine(line string, commentLinePrefix string) bool {
	var isCommentLine = false
	if IsValidString(commentLinePrefix) {
		columnNamesPrefixPattern, _ := regexp.Compile(commentLinePrefix)
		columnNamesPrefixMatcher := columnNamesPrefixPattern.FindString(line)

		if columnNamesPrefixMatcher != "" {
			return true
		} else {
			return false
		}
	}
	return isCommentLine
}

func DoesLineContainPrefix(line string, prefix string) bool {
	var doesLineContainPrefix = false
	pattern, _ := regexp.Compile(prefix)
	matcher := pattern.FindString(line)

	if matcher != "" {
		doesLineContainPrefix = true
	} else {
		doesLineContainPrefix = false
	}
	return doesLineContainPrefix
}

func RemovePrefixFromLine(line string, prefix string) string {
	columnNamesPrefixPattern, _ := regexp.Compile(prefix)
	columnNamesPrefixMatcher := columnNamesPrefixPattern.FindString(line)
	if columnNamesPrefixMatcher != "" {
		line = string(columnNamesPrefixMatcher[len(columnNamesPrefixMatcher)])
	}
	return line
}

func FindDelimiterByParsingPlotColumnPointsLine(line string) string {
	/*
	 * Right now, we are supporting two type of file contents
	 * One contains decimal numbers with . as decimal seperator and
	 * , as delimiter. Another contains , as decimal seperator and
	 * ; as delimiter
	 * So if the line contains ; its german csv file, else we will
	 * consider as english csv file
	 */

	if strings.Contains(line, GERMAN_CSV_FILE_DELIMITER) {
		return GERMAN_CSV_FILE_DELIMITER
	} else {
		return ENGLISH_CSV_FILE_DELIMITER
	}
}

func PopulatePlotPointsData(mapColumnPoints map[string][]string,
	arrColumnNames []string, arrColumnDataPoints []string, count int, numberLocale datamodel.NumberLocale) {
	var columnPoints []string
	for key, _ := range mapColumnPoints {
		columnPoints = mapColumnPoints[key]
		if numberLocale.Language != "" {
			var s string
			if numberLocale.Language == "de" {
				p := message.NewPrinter(language.German)
				s = p.Sprintf("%d\n", arrColumnDataPoints[findColumnNameIndexInArray(
					arrColumnNames, key)])
				columnPoints = append(columnPoints, s)
			} else {
				// p := message.NewPrinter(language.numberLocale.Language)
				// s = p.Sprintf("%d\n", arrColumnDataPoints[findColumnNameIndexInArray(
				// 	arrColumnNames, key)])
				columnPoints = append(columnPoints, arrColumnDataPoints[findColumnNameIndexInArray(
					arrColumnNames, key)])

			}

		} else {
			columnPoints = append(columnPoints, arrColumnDataPoints[findColumnNameIndexInArray(
				arrColumnNames, key)])
		}
		/*
		 * Remove existing records in case query is
		 * last n records.
		 */
		if count != 0 && len(columnPoints) > count {
			columnPoints = RemoveIndex(columnPoints, 0)
		}
		mapColumnPoints[key] = columnPoints
	}

}

func findColumnNameIndexInArray(arrColumnNames []string, columnName string) int {
	var index = 0
	for i := 0; i < len(arrColumnNames); i++ {
		if arrColumnNames[i] == columnName {
			index = i
			break
		}
	}
	return index
}

func RemoveIndex(s []string, index int) []string {
	ret := make([]string, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}
