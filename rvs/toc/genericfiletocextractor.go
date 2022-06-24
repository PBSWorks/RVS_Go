package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func GenericFileTOCExtractor(rvpProcessDataModel RVPProcessDataModel) datamodel.RVPPlotCType {
	var parsingStrategyModel = RVPParsingStrategyFinder(rvpProcessDataModel.RvpFileModel, rvpProcessDataModel.RvpResultFilePath)
	var commentPrefix = parsingStrategyModel.CommentsParserModel.Prefix
	var columnNamesLinePrefix = parsingStrategyModel.ColumnNamesParserModel.Prefix
	var columnNamesLineDelimiter = parsingStrategyModel.ColumnNamesParserModel.Delimiter
	var foundColumnNamesLine = false
	var simulationCount = 0

	file, err := os.Open(rvpProcessDataModel.RvpResultFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var rvpPlot datamodel.RVPPlotCType
	var columnNameBuilder string

	rvpPlot.RvpPlotColumnInfo.PlotName = common.UNTITLED_PLOT_NAME
	for scanner.Scan() {

		var line = strings.TrimSpace(scanner.Text())

		/*
		* Do not consider lines which are empty
		 */
		if !common.IsValidString(line) {
			continue
		} else if isCommentLine(line, commentPrefix) {
			continue
		} else {
			if foundColumnNamesLine {
				simulationCount++
			} else if common.IsValidString(columnNamesLinePrefix) {
				if doesLineContainPrefix(line, columnNamesLinePrefix) {
					line = removePrefixFromLine(line, columnNamesLinePrefix)
					columnNameBuilder = columnNameBuilder + line
				} else {
					foundColumnNamesLine = true
					populateColumnNames(columnNameBuilder, columnNamesLineDelimiter, rvpPlot)
					simulationCount++
				}
			} else {
				columnNameBuilder = columnNameBuilder + line
				foundColumnNamesLine = true
				var arrCurveNames = populateColumnNames(line, columnNamesLineDelimiter, rvpPlot)
				for i := 0; i < len(arrCurveNames); i++ {
					fmt.Println("arrCurveNames", arrCurveNames[i])
					rvpPlot.RvpPlotColumnInfo.ColumnNames = append(rvpPlot.RvpPlotColumnInfo.ColumnNames, strings.TrimSpace(arrCurveNames[i]))
				}
			}

		}

	}
	/*
	* File reading finished, add the rvp plot
	 */
	rvpPlot.Simulations.Count = simulationCount
	rvpPlot.Simulations.StartIndex = 1
	return rvpPlot
}

func RVPParsingStrategyFinder(rvpFileModel datamodel.RVPFileModel, resultFilePath string) datamodel.FileParsingStrategyModel {
	var fileName = common.GetFileName(resultFilePath)
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
		//	var numberFormat = fileParsingStrategyModel.DataPointsParserModel.NumberLocale

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
			if !common.IsValidString(line) || isCommentLine(line, commentLinePrefix) {
				continue
			} else if foundColumnNamesLine && foundDataPointsLine {

				if columnNamesCount == dataPointsCount && isDataPointsValid(line, dataPointsDelimiter, "") {
					correctFileParsingStrategyModel = fileParsingStrategyModel
					break
				} else {
					log.Println("Parsing strategy having id " + fileParsingStrategyModel.Id + " is not correct for file name " + fileName)
					break
				}
			} else {
				if foundColumnNamesLine {
					dataPointsCount = dataPointsCount + getDataPointsCount(line, dataPointsLinePrefix, dataPointsDelimiter)
					foundDataPointsLine = true
				} else if common.IsValidString(columnNamesLinePrefix) {
					if doesLineContainPrefix(line, columnNamesLinePrefix) {
						line = removePrefixFromLine(line, columnNamesLinePrefix)
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
				log.Println("Parsing strategy having id " + fileParsingStrategyModel.Id + "is not matching for file name " + fileName)
				break
			}

		}

	}
	return correctFileParsingStrategyModel
}

func isCommentLine(line string, commentLinePrefix string) bool {
	var isCommentLine = false
	if common.IsValidString(commentLinePrefix) {
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

func isDataPointsValid(line string, delimiter string, numberFormat string) bool {
	fmt.Println("delimiter", delimiter)
	var arrDataPoints = breakStringWithDelimiter(line, delimiter)
	if len(arrDataPoints) != 0 {
		for i := 0; i < len(arrDataPoints); i++ {
			if numberFormat == "" {
				_, err := strconv.Atoi(arrDataPoints[i])
				if err != nil {
					return false
				}
			} else {
				//	numberFormat.parse(dataPoint)
			}
		}
	} else {
		return false
	}
	return true
}

func breakStringWithDelimiter(line string, delimiter string) []string {
	pattern, _ := regexp.Compile(delimiter)
	matcher := pattern.FindString(line)
	fmt.Println("matcher", matcher)
	var lstArguments []string
	if matcher != "" {
		var arrArguments = pattern.Split(line, -1)
		for i := 0; i < len(arrArguments); i++ {
			if common.IsValidString(arrArguments[i]) {
				lstArguments = append(lstArguments, strings.TrimSpace(arrArguments[i]))
			}
		}
	}
	return lstArguments
}

func getDataPointsCount(line string, dataPointsLinePrefix string, dataPointsDelimiter string) int {
	var dataPointsCount = 0
	if common.IsValidString(dataPointsLinePrefix) {
		if doesLineContainPrefix(line, dataPointsLinePrefix) {
			line = removePrefixFromLine(line, dataPointsLinePrefix)
			dataPointsCount = dataPointsCount + getArgumentCountBasedOnDelimter(line, dataPointsDelimiter)
		}
	} else {
		dataPointsCount = dataPointsCount + getArgumentCountBasedOnDelimter(line, dataPointsDelimiter)
	}
	return dataPointsCount
}

func doesLineContainPrefix(line string, prefix string) bool {
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

func removePrefixFromLine(line string, prefix string) string {
	columnNamesPrefixPattern, _ := regexp.Compile(prefix)
	columnNamesPrefixMatcher := columnNamesPrefixPattern.FindString(line)
	if columnNamesPrefixMatcher != "" {
		line = string(columnNamesPrefixMatcher[len(columnNamesPrefixMatcher)])
	}
	return line
}

func getArgumentCountBasedOnDelimter(line string, delimiter string) int {

	pattern, _ := regexp.Compile(delimiter)
	arrArgument := pattern.Split(line, -1)
	var length = 0
	for i := 0; i < len(arrArgument); i++ {
		if common.IsValidString(arrArgument[i]) {
			length++
		}
	}
	return length
}

func populateColumnNames(line string, delimiter string, rvpPlot datamodel.RVPPlotCType) []string {
	fmt.Println("delimiter", delimiter)
	var arrCurveNames = breakStringWithDelimiter(line, delimiter)
	fmt.Println()
	return arrCurveNames
}
