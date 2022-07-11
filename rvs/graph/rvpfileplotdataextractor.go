package graph

import (
	"altair/rvs/datamodel"
	"altair/rvs/exception"
	"altair/rvs/utils"
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
		if !utils.IsValidString(line) || strings.HasPrefix(line, utils.COMMENT_STARTER) {
			continue
		}
		if beginPlotCount == 0 && rvpPlotDataModel.PlotName == line {
			beginPlotCount++
		} else if beginPlotCount == 3 && utils.END_PLOT == line {
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
				delimiter = utils.FindDelimiterByParsingPlotColumnPointsLine(line)
				if delimiter == utils.GERMAN_CSV_FILE_DELIMITER {
					numberLocale.Language = "de"
				}
				arrColumnNames = utils.BreakStringWithDelimiter(
					columnNames, delimiter)
				for i := 0; i < len(arrColumnNames); i++ {
					arrColumnNames[i] = strings.TrimSpace(arrColumnNames[i])
				}
			default:
				counter++
				if rowToBePickedIndex == counter {
					arrColumnPoints = utils.BreakStringWithDelimiter(line, delimiter)
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
						utils.PopulatePlotPointsData(rvpPlotDataModel.MapColumnPoints,
							arrColumnNames, arrColumnPoints, count, numberLocale)
						rowToBePickedIndex = rowToBePickedIndex + step
						if endIndex != utils.SIMULATION_END_INDEX && rowToBePickedIndex > endIndex {
							finishedPlotDataReading = true
						}
					}

				}

			}
		}
	}
	return rvpPlotDataModel, nil
}
