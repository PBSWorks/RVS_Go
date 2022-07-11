package common

import (
	"altair/rvs/datamodel"
	"altair/rvs/utils"
	"bufio"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	l "altair/rvs/globlog"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func RVPParsingStrategyFinder(rvpFileModel datamodel.RVPFileModel, resultFilePath string) datamodel.FileParsingStrategyModel {
	var fileName = utils.GetFileName(resultFilePath)
	var correctFileParsingStrategyModel datamodel.FileParsingStrategyModel
	var ListFileParsingStrategyModelData = rvpFileModel.FileParsingStrategiesModel.ListFileParsingStrategyModel
	for i := 0; i < len(ListFileParsingStrategyModelData); i++ {
		var fileParsingStrategyModel = rvpFileModel.FileParsingStrategiesModel.ListFileParsingStrategyModel[i]
		var columnNamesCount = 0
		var dataPointsCount = 0
		var foundColumnNamesLine = false
		var foundDataPointsLine = false

		var commentLinePrefix = fileParsingStrategyModel.CommentsParserModel.Prefix
		var columnNamesLinePrefix = fileParsingStrategyModel.ColumnNamesParserModel.Prefix
		var dataPointsLinePrefix = fileParsingStrategyModel.DataPointsParserModel.Prefix
		var columnNamesDelimiter = fileParsingStrategyModel.ColumnNamesParserModel.Delimiter
		var dataPointsDelimiter = fileParsingStrategyModel.DataPointsParserModel.Delimiter
		var numberLocale = fileParsingStrategyModel.DataPointsParserModel.NumberLocale

		file, err := os.Open(resultFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {

			// remove any unnecessary whitespaces
			var line = strings.TrimSpace(scanner.Text())
			/*
			 * Do not consider lines which are empty or contains comments
			 */
			if !utils.IsValidString(line) || utils.IsCommentLine(line, commentLinePrefix) {
				continue
			} else if foundColumnNamesLine && foundDataPointsLine {

				if columnNamesCount == dataPointsCount && IsDataPointsValid(line, dataPointsDelimiter, numberLocale) {
					correctFileParsingStrategyModel = fileParsingStrategyModel
					break
				} else {
					l.Log().Info("Parsing strategy having id " + fileParsingStrategyModel.Id + " is not correct for file name " + fileName)
					break
				}
			} else {
				if foundColumnNamesLine {
					dataPointsCount = dataPointsCount + getDataPointsCount(line, dataPointsLinePrefix, dataPointsDelimiter)
					foundDataPointsLine = true
				} else if utils.IsValidString(columnNamesLinePrefix) {
					if utils.DoesLineContainPrefix(line, columnNamesLinePrefix) {
						line = utils.RemovePrefixFromLine(line, columnNamesLinePrefix)
						columnNamesCount = columnNamesCount + getArgumentCountBasedOnDelimter(line, columnNamesDelimiter)
					} else {
						foundColumnNamesLine = true
						dataPointsCount = dataPointsCount + getDataPointsCount(line, dataPointsLinePrefix, dataPointsDelimiter)
						foundDataPointsLine = true
					}
				} else {
					columnNamesCount = columnNamesCount + getArgumentCountBasedOnDelimter(line, columnNamesDelimiter)
					foundColumnNamesLine = true
				}
			}
		}
		if correctFileParsingStrategyModel.ColumnNamesParserModel.Delimiter != "" {
			break
		} else {
			if foundColumnNamesLine && foundDataPointsLine {
				correctFileParsingStrategyModel = fileParsingStrategyModel
				//break
			} else {
				l.Log().Info("Parsing strategy having id " + fileParsingStrategyModel.Id + "is not matching for file name " + fileName)
				break
			}

		}

	}
	return correctFileParsingStrategyModel
}

func IsDataPointsValid(line string, delimiter string, numberLocale datamodel.NumberLocale) bool {
	var arrDataPoints = utils.BreakStringWithDelimiter(line, delimiter)
	if len(arrDataPoints) != 0 {
		for i := 0; i < len(arrDataPoints); i++ {
			if numberLocale.Language == "" {
				_, err := strconv.Atoi(arrDataPoints[i])
				if err != nil {
					return false
				}
			} else {
				if numberLocale.Language == "de" {
					p := message.NewPrinter(language.German)
					s := p.Sprintf("%d\n", arrDataPoints[i])
					if s == "" {
						return false
					}

				}
			}
		}
	} else {
		return false
	}
	return true
}

func getDataPointsCount(line string, dataPointsLinePrefix string, dataPointsDelimiter string) int {
	var dataPointsCount = 0
	if utils.IsValidString(dataPointsLinePrefix) {
		if utils.DoesLineContainPrefix(line, dataPointsLinePrefix) {
			line = utils.RemovePrefixFromLine(line, dataPointsLinePrefix)
			dataPointsCount = dataPointsCount + getArgumentCountBasedOnDelimter(line, dataPointsDelimiter)
		}
	} else {
		dataPointsCount = dataPointsCount + getArgumentCountBasedOnDelimter(line, dataPointsDelimiter)
	}
	return dataPointsCount
}

func getArgumentCountBasedOnDelimter(line string, delimiter string) int {

	pattern, _ := regexp.Compile(delimiter)
	arrArgument := pattern.Split(line, -1)
	var length = 0
	for i := 0; i < len(arrArgument); i++ {
		if utils.IsValidString(arrArgument[i]) {
			length++
		}
	}
	return length
}
