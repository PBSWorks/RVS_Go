package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"bufio"
	"log"
	"os"
	"strings"
)

func RVPFileTOCExtractor(sRVPFilePath string) datamodel.RVPPlotCType {

	var rvpPlot datamodel.RVPPlotCType
	var beginPlotCount = 0
	var simulationCount = 0
	var arrCurveNamesLine = ""

	file, err := os.Open(sRVPFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var line = strings.TrimSpace(scanner.Text())
		//Logging RVP Version
		if strings.HasPrefix(line, common.RVP_VERSION_SYMBOL) {
			log.Println("Parsing RVP file with Version -> " + line)
		}
		/*
		* Do not consider lines which are either empty or contains comments
		 */
		if !common.IsValidString(line) || strings.HasPrefix(line, common.COMMENT_STARTER) {
			continue
		}
		/*
		* Header telling plot description has begin
		 */
		if beginPlotCount == 0 && strings.EqualFold(common.BEGIN_PLOT, line) {
			beginPlotCount++
		} else if beginPlotCount == 4 && strings.EqualFold(common.END_PLOT, line) {
			rvpPlot.Simulations.Count = simulationCount
			rvpPlot.Simulations.StartIndex = 1

			// rvpPlot = null
			// simulations = null
			// beginPlotCount = 0
			// simulationCount = 0
		} else {
			switch beginPlotCount {
			/*
			 * Receiving plot name information
			 */
			case 1:
				rvpPlot.RvpPlotColumnInfo.PlotName = line
				beginPlotCount++
			/*
			 * Receiving plot column names information
			 */
			case 2:
				beginPlotCount++
				arrCurveNamesLine = line
				/*
				 * We do not know the correct delimiter at this point of
				 * time, next line will only contain plot points information,
				 * finding right delimiter from that line will be more decisive.
				 */
			case 3:
				simulationCount++
				beginPlotCount++
				var delimiter = common.FindDelimiterByParsingPlotColumnPointsLine(line)
				var arrCurveNames = common.BreakStringWithDelimiter(arrCurveNamesLine, delimiter)
				for i := 0; i < len(arrCurveNames); i++ {
					rvpPlot.RvpPlotColumnInfo.ColumnNames = append(rvpPlot.RvpPlotColumnInfo.ColumnNames, strings.TrimSpace(arrCurveNames[i]))
				}
			default:
				simulationCount++
			}
		}
	}

	return rvpPlot
}
