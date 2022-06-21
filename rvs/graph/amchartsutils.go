package graph

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var m_bIsXNegative bool = false

var m_bIsYNegative bool = false

/**
 * Reference to the Precision value
 */
const PREF_PRECISION = "decimalPrecision"

/**
 * Reference to the Default Decimal Precision
 */
const DEFAULT_PRECISION = 8

/**
 * Reference to Hash symbol
 */
const HASH = "#"

/**
 * Data File Name
 */
const FILE_NAME_DATA_ORIGINAL = "Data_Original.csv"

/**
 * Data File Name
 */
const FILE_NAME_DATA = "Data.csv"

/**
 * Log Data File Name
 */
const FILE_NAME_DATA_LOGX = "Data_LogX.csv"

/**
 * Log Data File Name
 */
const FILE_NAME_DATA_LOGY = "Data_LogY.csv"

/**
 * Log Data File Name
 */
const FILE_NAME_DATA_LOGX_LOGY = "Data_LogX_LogY.csv"

/**
 * Chart File Name
 */
const FILE_NAME_CHART_HTML = "Chart.html"

const SMALL_POSITIVE = "0.0000001"

const MAX_DATA_POINT_LIMIT = 50000

var m_formatLstDoubleData [][]float64

/**
 * Creates XML data file with the given details.
 *
 * @param sDirPath
 * @param lstDoubleData
 * @param sXAxisName
 * @param sYAxisName
 * @param lstCurveNames
 * @param iPrecisionRange
 * @throws Exception
 */
func createCSVDataFiles(sDirPath string, lstDoubleData [][]float64,
	sXAxisName string, sYAxisName string, lstCurveNames []string, iPrecisionRange int) {
	// Iterate the list of Double Data
	m_formatLstDoubleData = nil
	for i := 0; i < len(lstDoubleData); i++ {
		var formatListData []float64
		var listData = lstDoubleData[i]
		for j := 0; j < len(listData); j++ {
			if i == 0 {
				if listData[j] < 0 {
					m_bIsXNegative = true
				}
			} else {
				if listData[j] < 0 {
					m_bIsYNegative = true
				}
			}

			s := fmt.Sprintf("%.8f", listData[j])
			if data, err := strconv.ParseFloat(s, 64); err == nil {
				formatListData = append(formatListData, data)
			}
		}
		m_formatLstDoubleData = append(m_formatLstDoubleData, formatListData)
	}
	var filepath = sDirPath + "/" + FILE_NAME_DATA_ORIGINAL

	f, err := os.Create(filepath)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	writeOriginalChartData(lstDoubleData, sXAxisName, lstCurveNames, f)

}

func writeOriginalChartData(lstDoubleData [][]float64, sAxisName string,
	lstCurveNames []string, f *os.File) {
	_, err2 := f.WriteString(sAxisName)

	if err2 != nil {
		log.Fatal(err2)
	}

	for _, sCurveName := range lstCurveNames {

		_, err2 := f.WriteString("," + sCurveName)

		if err2 != nil {
			log.Fatal(err2)
		}
	}

	_, err3 := f.WriteString("\n")

	if err3 != nil {
		log.Fatal(err2)
	}

	for i := 0; i < len(lstDoubleData[0]); i++ {
		var sVal string = ""
		for j := 0; j < len(lstDoubleData); j++ {
			var listVal = lstDoubleData[j]

			// this is the X list
			if i < len(listVal) {
				if j == 0 {
					lstData := fmt.Sprintf("%.8f", listVal[i])
					sVal = sVal + lstData
				} else {
					lstData := fmt.Sprintf("%.8f", listVal[i])
					sVal = sVal + "," + lstData
				}
			} else {
				sVal = sVal + ","
			}
		}

		_, err2 := f.WriteString(sVal + "\n")

		if err2 != nil {
			log.Fatal(err2)
		}

	}
}

func getNumberOfDataPoints() int {
	var index = 0
	var flag = false
	for i := 0; i < len(m_formatLstDoubleData); i++ {
		if flag {
			index = index + len(m_formatLstDoubleData[i])
		}
		flag = true
	}
	return index
}

func getDataPoints(isLogX bool, isLogY bool) string {

	var flag = false
	var outerArray strings.Builder

	outerArray.WriteString("[")
	for n := 0; n < len(m_formatLstDoubleData); n++ {
		var innerArray strings.Builder
		var yPointList = m_formatLstDoubleData[n]
		for i := 0; i < len(yPointList); i++ {
			var tempYVal string
			if len(yPointList) > 0 {
				tempYVal = fmt.Sprintf("%.8f", yPointList[i])
			} else {
				tempYVal = "NAN"
			}
			innerArray.WriteString(tempYVal)
			if len(m_formatLstDoubleData[n]) != (i + 1) {
				innerArray.WriteString(",")
			}

		}
		if flag {
			outerArray.WriteString(",{")
		} else {
			outerArray.WriteString("{")
			flag = true
		}
		if n == 0 {
			outerArray.WriteString("x")
			outerArray.WriteString("1")
		} else {
			outerArray.WriteString("y")
			outerArray.WriteString(strconv.Itoa(n))
		}

		outerArray.WriteString(":")
		outerArray.WriteString("[")
		outerArray.WriteString(innerArray.String())
		outerArray.WriteString("]")
		outerArray.WriteString("}")
		// outerArray.append(",");

	}
	// outerArray.append("}");

	outerArray.WriteString("]")

	return outerArray.String()
}

func getDataPointsLogX() string {
	var flag = false
	var outerArray strings.Builder
	outerArray.WriteString("[")
	for n := 0; n < len(m_formatLstDoubleData); n++ {
		var innerArray strings.Builder
		var yPointList = m_formatLstDoubleData[n]

		for i := 0; i < len(yPointList); i++ {

			var tempYVal string
			if len(yPointList) > 0 {
				if n == 0 {
					if len(yPointList) > 0 {
						if yPointList[i] <= 0 {
							tempYVal = SMALL_POSITIVE
						} else {
							tempYVal = fmt.Sprintf("%.8f", yPointList[i])
						}
					} else {
						tempYVal = "NAN"
					}
				} else {
					tempYVal = fmt.Sprintf("%.8f", yPointList[i])
				}
			} else {
				tempYVal = "NAN"
			}
			innerArray.WriteString(tempYVal)
			if len(m_formatLstDoubleData[n]) != (i + 1) {
				innerArray.WriteString(",")
			}

		}
		if flag {
			outerArray.WriteString(",{")
		} else {
			outerArray.WriteString("{")
			flag = true
		}
		if n == 0 {
			outerArray.WriteString("x")
			outerArray.WriteString(strconv.Itoa(1))
		} else {
			outerArray.WriteString("y")
			outerArray.WriteString(strconv.Itoa(n))
		}

		outerArray.WriteString(":")
		outerArray.WriteString("[")
		outerArray.WriteString(innerArray.String())
		outerArray.WriteString("]")
		outerArray.WriteString("}")
		// outerArray.append(",");

	}
	// outerArray.append("}");

	outerArray.WriteString("]")

	return outerArray.String()
}

func getDataPointsLogY() string {
	var flag = false
	var outerArray strings.Builder
	outerArray.WriteString("[")
	for n := 0; n < len(m_formatLstDoubleData); n++ {
		var innerArray strings.Builder

		for i := 0; i < len(m_formatLstDoubleData[n]); i++ {
			var yPointList = m_formatLstDoubleData[n]
			var tempYVal string

			if len(yPointList) > 0 {
				if n != 0 {
					if len(yPointList) > 0 {
						if yPointList[i] <= 0 {
							tempYVal = SMALL_POSITIVE
						} else {
							tempYVal = fmt.Sprintf("%.8f", yPointList[i])
						}
					} else {
						tempYVal = "NAN"
					}
				} else {
					tempYVal = fmt.Sprintf("%.8f", yPointList[i])
				}

			} else {
				tempYVal = "NAN"
			}
			innerArray.WriteString(tempYVal)
			if len(m_formatLstDoubleData[n]) != (i + 1) {
				innerArray.WriteString(",")
			}

		}
		if flag {
			outerArray.WriteString(",{")
		} else {
			outerArray.WriteString("{")
			flag = true
		}
		if n == 0 {
			outerArray.WriteString("x")
			outerArray.WriteString(strconv.Itoa(1))
		} else {
			outerArray.WriteString("y")
			outerArray.WriteString(strconv.Itoa(n))
		}

		outerArray.WriteString(":")
		outerArray.WriteString("[")
		outerArray.WriteString(innerArray.String())
		outerArray.WriteString("]")
		outerArray.WriteString("}")
		// outerArray.append(",");

	}
	// outerArray.append("}");

	outerArray.WriteString("]")
	return outerArray.String()
}

func getDataPointsLogYandLogX() string {
	var flag = false
	var outerArray strings.Builder
	outerArray.WriteString("[")
	for n := 0; n < len(m_formatLstDoubleData); n++ {
		var innerArray strings.Builder

		for i := 0; i < len(m_formatLstDoubleData[n]); i++ {
			var yPointList = m_formatLstDoubleData[n]
			var tempYVal string

			if len(yPointList) > 0 {

				if len(yPointList) > 0 {
					if yPointList[i] <= 0 {
						tempYVal = SMALL_POSITIVE
					} else {
						tempYVal = fmt.Sprintf("%.8f", yPointList[i])
					}
				} else {
					tempYVal = "NAN"
				}
			} else {
				tempYVal = "NAN"
			}
			innerArray.WriteString(tempYVal)
			if len(m_formatLstDoubleData[n]) != (i + 1) {
				innerArray.WriteString(",")
			}

		}
		if flag {
			outerArray.WriteString(",{")
		} else {
			outerArray.WriteString("{")
			flag = true
		}
		if n == 0 {
			outerArray.WriteString("x")
			outerArray.WriteString(strconv.Itoa(1))
		} else {
			outerArray.WriteString("y")
			outerArray.WriteString(strconv.Itoa(n))
		}

		outerArray.WriteString(":")
		outerArray.WriteString("[")
		outerArray.WriteString(innerArray.String())
		outerArray.WriteString("]")
		outerArray.WriteString("}")
		// outerArray.append(",");

	}
	// outerArray.append("}");

	outerArray.WriteString("]")
	return outerArray.String()
}

func doesXCurveContainsNegativeValues() bool {
	return m_bIsXNegative
}
func doesYCurveContainsNegativeValues() bool {
	return m_bIsXNegative
}
