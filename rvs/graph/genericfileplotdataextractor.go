package graph

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"bufio"
	"log"
	"os"
	"strings"
)

var foundrvpColumnNamesLine = false
var rowToBePickedIndex = 0
var endIndex = 0
var step = 0
var count = 0
var arrColumnNames []string
var columnNames = ""
var commentPrefix = ""
var columnNamesLinePrefix = ""
var columnNamesLineDelimiter = ""
var dataPointsDelimiter = ""
var numberLocale datamodel.NumberLocale
var finishedPlotDataReading = false

func GenericFilePlotDataExtractor(RvpResultFilePath string,
	rvpProcessDataModel RVPProcessDataModel, rvpPlotDataModel RVPPlotDataModel, tempSimulationQuery TemporarySimulationQuery) RVPPlotDataModel {
	rowToBePickedIndex = tempSimulationQuery.StartIndex
	endIndex = tempSimulationQuery.EndIndex
	step = tempSimulationQuery.Step
	count = tempSimulationQuery.Count
	var counter = 0
	var columnNameBuilder strings.Builder

	var parsingStrategyModel = common.RVPParsingStrategyFinder(rvpProcessDataModel.RvpFileModel, rvpProcessDataModel.RvpResultFilePath)
	numberLocale = parsingStrategyModel.DataPointsParserModel.NumberLocale

	commentPrefix = parsingStrategyModel.CommentsParserModel.Prefix
	columnNamesLinePrefix = parsingStrategyModel.ColumnNamesParserModel.Prefix
	columnNamesLineDelimiter = parsingStrategyModel.ColumnNamesParserModel.Delimiter
	dataPointsDelimiter = parsingStrategyModel.DataPointsParserModel.Delimiter

	file, err := os.Open(rvpProcessDataModel.RvpResultFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() && !finishedPlotDataReading {
		var line = strings.TrimSpace(scanner.Text())
		if !common.IsValidString(line) || strings.HasPrefix(line, common.COMMENT_STARTER) {
			continue
		} else if common.IsCommentLine(line, commentPrefix) {
			continue
		}
		if foundrvpColumnNamesLine {
			counter++
			populateDataPoints(counter, line, rvpPlotDataModel)
		} else if common.IsValidString(columnNamesLinePrefix) {
			if common.DoesLineContainPrefix(line, columnNamesLinePrefix) {
				line = common.RemovePrefixFromLine(line, columnNamesLinePrefix)
				columnNameBuilder.WriteString(line)
			} else {
				foundrvpColumnNamesLine = true
				columnNames = columnNameBuilder.String()
				arrColumnNames = common.BreakStringWithDelimiter(columnNames, columnNamesLineDelimiter)
				counter++
				populateDataPoints(counter, line, rvpPlotDataModel)
			}
		} else {
			foundrvpColumnNamesLine = true
			columnNames = line
			arrColumnNames = common.BreakStringWithDelimiter(columnNames, columnNamesLineDelimiter)

		}

	}
	return rvpPlotDataModel
}

func populateDataPoints(counter int, line string, rvpPlotDataModel RVPPlotDataModel) error {
	var arrColumnPoints []string
	if rowToBePickedIndex == counter {
		arrColumnPoints = common.BreakStringWithDelimiter(line, dataPointsDelimiter)

		for i := 0; i < len(arrColumnPoints); i++ {
			arrColumnPoints[i] = strings.TrimSpace(arrColumnPoints[i])
		}
		if len(arrColumnNames) != len(arrColumnPoints) {
			var sb strings.Builder
			sb.WriteString("The number of points are not equal to the number of columns in the file")
			sb.WriteString("\n")
			sb.WriteString("Column Names")
			sb.WriteString("\n")
			sb.WriteString(columnNames)
			sb.WriteString("\n")
			sb.WriteString("Point Names")
			sb.WriteString("\n")
			sb.WriteString(line)

			return &exception.RVSError{
				Errordetails: "",
				Errorcode:    "11000",
				Errortype:    "TYPE_QUERY_FAILED",
			}
		} else {
			common.PopulatePlotPointsData(rvpPlotDataModel.MapColumnPoints, arrColumnNames, arrColumnPoints, count, numberLocale)
			rowToBePickedIndex = rowToBePickedIndex + step
			if endIndex != common.SIMULATION_END_INDEX && rowToBePickedIndex > endIndex {
				finishedPlotDataReading = true
			}
		}
	}
	return nil
}
