package utils

import (
	"altair/rvs/datamodel"
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const FORWARD_SLASH = "/"
const BACK_SLASH = "\\"
const DOUBLE_QUOTES = "\""
const SINGLE_QUOTE = "'"
const NEWLINE = "\n"
const TEMP_FILE_NAME = "tempfile"
const RM_TOC_XML_FILES = "/RM_TOC_XML_FILES"
const RM_SCRIPT_FILES = "/RM_SCRIPT_FILES"
const RM_OUTPUT_FILES = "/RM_OUTPUT_FILES"
const RM_DOWNLOADS = "/RM_DOWNLOADS"
const COMPOSE_PRODUCT_ID = "COMPOSE"
const HYPERWORKS_PRODUCT_ID = "ALTAIR_HYPERWORKS"
const Siteconfigfile = `/site_config.xml`
const MAX_COMPOSE_INSTANCE_COUNT = 500
const COMPOSE_WINDOWS_64BIT_EXEC = "/Compose.bat"
const COMPOSE_UNIX_EXEC = "/scripts/Compose_Batch"
const PLOT_OML_PATH = "/scripts"
const TEMP_OML_FILE_NAME = "temp.oml"

const ROOT_ASSEMBLY = "ROOT"
const SCRIPTS = "/scripts"
const PLOT_TOC_OUTPUT_FILE_NAME_PART = "PlotTOC.json"
const PLOT_TOC_OUTPUT_OML_NAME_PART = "PlotTOC.oml"
const PLOT_TOC_OML_FILE_NAME = "GetPlotTOC.oml"
const MODEL_COMPONENTS_FILE_NAME = "ModelComponents.json"
const MODEL_COMP_SOURCE_FILE_NAME = "/ModelComponents.cfg"
const MODEL_COMP_CFG_FILE_NAME = "/GetModelComps.cfg"
const PLOT_GRAPH_OML_FILE_NAME = "GetPlotData.oml"
const ANIM_TOC_OUTPUT_FILE_NAME = "AnimTOC.json"
const ANIM_TOC_CFG_PATH = "/GetAnimationTOC.cfg"
const MODEL_FILE_EXT = ".model"

const ANIM_TOC_MARKER_TAG = "$#$HWE_RS_ANIM_TOC_FILE$#$"
const MODEL_COMPS_SCRIPT_PATH_TAG = "$#$MODEL_COMPS_SCRIPT_PATH$#$"
const TEMP_MODEL_FILE_TAG = "$#$HWE_RS_TEMP_MODEL_FILE$#$"

const COMP_LIST_FILE_PATH = "/components.txt"

const STATISTICS_TAG = "Statistics"
const STATISTIC_TAG = "Statistic"
const ASSEMBLIES = "assemblies"
const NODES = "nodes"
const PARTS = "parts"
const SYSTEMS = "systems"
const ELEMENTS = "elements"
const POOLS_TAG = "Pools"
const ASSEMBLY_TAG = "Assembly"
const PARTS_TAG = "Parts"
const PART_TAG = "Part"
const ASSEMBLY_POOL_TAG = "AssemblyPool"
const ELEMENT_POOL_TAG = "ElementPool"
const NODE_POOL_TAG = "NodePool"
const PART_POOL_TAG = "PartPool"
const SYSTEM_POOL_TAG = "SystemPool"
const NAME_ATTRIBUTE = "name"
const TYPE_ATTRIBUTE = "type"
const PAS_URL_VALUE = "/pas"
const JOB_OPERATION = "/joboperation"
const REST_SERVICE_URL = "/restservice"
const JOB_RUNNING_STATE = "R"
const PBS_JOB_RUNNING_STATE = "R"
const PBS_JOB_EXIT_STATE = "E"
const TEMPORARY_TEMPLATE_FILE_NAME = "Untitled.rvst"

var (
	SeriesVsBaseRegEx     = make(map[string]string)
	SeriesRegexVsWildcard = make(map[string]string)
	RvpFilesModel         datamodel.SupportedRVPFilesModel
)

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func GetRandomString(prefix string) string {
	rand.Seed(time.Now().UnixNano())
	return prefix + "_" + randomString(10) // print 10 chars
}
func GetUniqueRandomIntValue() int64 {
	uniqueNumber := time.Now().UnixNano() / (1 << 22)
	return uniqueNumber
}

func IsValidString(sText string) bool {
	return sText != ""
}

func IsWindows() bool {
	os := runtime.GOOS
	return os == "windows"
}

func Is32BitOS() bool {
	osarch := runtime.GOARCH
	return osarch == "32"
}

func GetRSHome() string {
	var rshome string = ""
	os := runtime.GOOS
	if os == "windows" {
		rshome = "/opt/go/files"
	} else {
		rshome = "/opt/go/files"
	}
	return rshome
}

func GetPlatformIndependentFilePath(sFilePath string, bHandleWhiteSpaces bool) string {
	var sPath string
	if IsValidString(sFilePath) {
		sPath = strings.Replace(sFilePath, BACK_SLASH, FORWARD_SLASH, -1)
	} else {
		return ""
	}

	// if bHandleWhiteSpaces {
	// 	if IsWindows() {
	// 		sPath = sPath
	// 	} else {
	// 		sPath = sPath
	// 	}
	// }
	return sPath
}

func GetFileNameWithoutExtension(sFilePath string) string {
	return strings.TrimSuffix(filepath.Base(sFilePath), filepath.Ext(sFilePath))
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}
func GetFileName(Filepath string) string {
	_, file := filepath.Split(Filepath)
	return file
}

func GetDirPath(Filepath string) string {
	dir, _ := filepath.Split(Filepath)
	return dir
}

/**
 * Breaks a give string based on the delimiter passed and returns
 * array of string containing the broken strings.
 * @param line
 * @param delimiter
 * @return
 */
// func BreakStringWithDelimiter(line string, delimiter string) []string {
// 	var arrStringTokens = strings.Split(line, delimiter)
// 	return arrStringTokens
// }

func BreakStringWithDelimiter(line string, delimiter string) []string {
	pattern, _ := regexp.Compile(delimiter)
	matcher := pattern.FindString(line)
	var lstArguments []string
	if matcher != "" {
		var arrArguments = pattern.Split(line, -1)
		for i := 0; i < len(arrArguments); i++ {
			if IsValidString(arrArguments[i]) {
				lstArguments = append(lstArguments, strings.TrimSpace(arrArguments[i]))
			}
		}
	}
	return lstArguments
}
func GetDataDirectoryPath(servername string, username string) string {
	var sDataDirectoryPath = getNewDirPath(servername, username)
	if err := os.MkdirAll(sDataDirectoryPath, 0755); err != nil {
		log.Fatal(err)
	}

	return sDataDirectoryPath
}

func getNewDirPath(sServerName string, username string) string {
	var sUniqueDirPath string = ""
	sUniqueDirPath = GetRMDataDirectory() + username + "/" + strconv.FormatInt(GetUniqueRandomIntValue(), 10)
	//log.Fatal("Error occured while creating new directory path")
	return sUniqueDirPath
}

func GetRMDataDirectory() string {
	//return os.Getenv("PBSWORKS_HOME") + "/data/resultmanager" + "/" + DATA_DIR_NAME + "/"

	// var dir = common.GetRSHome() + "/data/"
	// if err := os.Mkdir(dir, 0755); err != nil {
	// 	log.Fatal(err)
	// }
	return GetRSHome() + "/data/"

}
