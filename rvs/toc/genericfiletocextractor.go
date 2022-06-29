package toc

import (
	"altair/rvs/common"
	"altair/rvs/datamodel"
	"bufio"
	"log"
	"os"
	"strings"
)

func GenericFileTOCExtractor(rvpProcessDataModel RVPProcessDataModel) datamodel.RVPPlotCType {
	var parsingStrategyModel = common.RVPParsingStrategyFinder(rvpProcessDataModel.RvpFileModel, rvpProcessDataModel.RvpResultFilePath)
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
		} else if common.IsCommentLine(line, commentPrefix) {
			continue
		} else {
			if foundColumnNamesLine {
				simulationCount++
			} else if common.IsValidString(columnNamesLinePrefix) {
				if common.DoesLineContainPrefix(line, columnNamesLinePrefix) {
					line = common.RemovePrefixFromLine(line, columnNamesLinePrefix)
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

func populateColumnNames(line string, delimiter string, rvpPlot datamodel.RVPPlotCType) []string {
	var arrCurveNames = common.BreakStringWithDelimiter(line, delimiter)
	return arrCurveNames
}
