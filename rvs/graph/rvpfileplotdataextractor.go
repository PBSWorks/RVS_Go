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

var finishedRvpPlotDataReading = false

func RVPFilePlotDataExtractor(RvpResultFilePath string,
	rvpProcessDataModel RVPProcessDataModel, rvpPlotDataModel RVPPlotDataModel,
	tempSimulationQuery TemporarySimulationQuery) (RVPPlotDataModel, error) {
	var beginPlotCount = 0
	var columnNames = ""
	var delimiter = ""
	var arrColumnNames []string
	var arrColumnPoints []string
	var rowToBePickedIndex = tempSimulationQuery.StartIndex
	var endIndex = tempSimulationQuery.EndIndex
	var step = tempSimulationQuery.Step
	var count = tempSimulationQuery.Count
	var counter = 0
	var numberLocale datamodel.NumberLocale
	file, err := os.Open(RvpResultFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() && !finishedRvpPlotDataReading {
		var line = strings.TrimSpace(scanner.Text())
		if !common.IsValidString(line) || strings.HasPrefix(line, common.COMMENT_STARTER) {
			continue
		}
		if beginPlotCount == 0 && rvpPlotDataModel.PlotName == line {
			beginPlotCount++
		} else if beginPlotCount == 3 && common.END_PLOT == line {
			break
		} else if beginPlotCount > 0 {
			switch beginPlotCount {
			case 1:
				beginPlotCount++
				columnNames = line
			/*
			* Receiving plot column names information
			 */
			case 2:
				beginPlotCount++
				delimiter = common.FindDelimiterByParsingPlotColumnPointsLine(line)
				if delimiter == common.GERMAN_CSV_FILE_DELIMITER {
					numberLocale.Language = "de"
				}
				arrColumnNames = common.BreakStringWithDelimiter(
					columnNames, delimiter)
				for i := 0; i < len(arrColumnNames); i++ {
					arrColumnNames[i] = strings.TrimSpace(arrColumnNames[i])
				}
			default:
				counter++
				if rowToBePickedIndex == counter {
					arrColumnPoints = common.BreakStringWithDelimiter(line, delimiter)
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

						return RVPPlotDataModel{}, &exception.RVSError{
							Errordetails: "",
							Errorcode:    "11000",
							Errortype:    "TYPE_QUERY_FAILED",
						}

					} else {
						common.PopulatePlotPointsData(rvpPlotDataModel.MapColumnPoints,
							arrColumnNames, arrColumnPoints, count, numberLocale)
						rowToBePickedIndex = rowToBePickedIndex + step
						if endIndex != common.SIMULATION_END_INDEX && rowToBePickedIndex > endIndex {
							finishedPlotDataReading = true
						}
					}

				}

			}
		}
	}
	return rvpPlotDataModel, nil
}
