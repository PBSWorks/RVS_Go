package graph

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"altair/rvs/utils"
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
		if !utils.IsValidString(line) || strings.HasPrefix(line, utils.COMMENT_STARTER) {
			continue
		} else if utils.IsCommentLine(line, commentPrefix) {
			continue
		}
		if foundrvpColumnNamesLine {
			counter++
			populateDataPoints(counter, line, rvpPlotDataModel)
		} else if utils.IsValidString(columnNamesLinePrefix) {
			if utils.DoesLineContainPrefix(line, columnNamesLinePrefix) {
				line = utils.RemovePrefixFromLine(line, columnNamesLinePrefix)
				columnNameBuilder.WriteString(line)
			} else {
				foundrvpColumnNamesLine = true
				columnNames = columnNameBuilder.String()
				arrColumnNames = utils.BreakStringWithDelimiter(columnNames, columnNamesLineDelimiter)
				counter++
				populateDataPoints(counter, line, rvpPlotDataModel)
			}
		} else {
			foundrvpColumnNamesLine = true
			columnNames = line
			arrColumnNames = utils.BreakStringWithDelimiter(columnNames, columnNamesLineDelimiter)

		}

	}
	return rvpPlotDataModel
}

func populateDataPoints(counter int, line string, rvpPlotDataModel RVPPlotDataModel) error {
	var arrColumnPoints []string
	if rowToBePickedIndex == counter {
		arrColumnPoints = utils.BreakStringWithDelimiter(line, dataPointsDelimiter)

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
			utils.PopulatePlotPointsData(rvpPlotDataModel.MapColumnPoints, arrColumnNames, arrColumnPoints, count, numberLocale)
			rowToBePickedIndex = rowToBePickedIndex + step
			if endIndex != utils.SIMULATION_END_INDEX && rowToBePickedIndex > endIndex {
				finishedPlotDataReading = true
			}
		}
	}
	return nil
}
