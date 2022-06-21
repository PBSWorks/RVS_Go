package toc

import (
	"altair/rvs/common"
	"fmt"
	"os"
	"time"
)

var m_ScriptPath = ""

func ExecuteComposeApplicatopn(plottocomlfile string, username string, password string) {
	m_ScriptPath = plottocomlfile
	dt := time.Now()
	common.RunCommand(buildCommandArrayForOperatingSystem(), username, password)
	dt1 := time.Now()
	diff := dt1.Sub(dt)
	fmt.Println("Compose Execution Time: ", diff)
}

func buildCommandArrayForOperatingSystem() []string {
	lstOfCmdItems := []string{}
	var sExec string
	if common.IsWindows() {
		sExec = common.GetProductInstallationLocation(common.COMPOSE_PRODUCT_ID) + common.COMPOSE_WINDOWS_64BIT_EXEC
		lstOfCmdItems = append(lstOfCmdItems, common.GetPlatformIndependentFilePath(sExec, true))

	} else {
		sExec = common.GetProductInstallationLocation(common.COMPOSE_PRODUCT_ID) + common.COMPOSE_UNIX_EXEC
		lstOfCmdItems = append(lstOfCmdItems, common.GetPlatformIndependentFilePath(sExec, true))
	}
	info, err := os.Stat(sExec)
	if os.IsNotExist(err) {
		fmt.Println(info)
		fmt.Println("sExec: ", sExec)
		fmt.Println("Compose Execution file does not exists.")
	}
	lstOfCmdItems = append(lstOfCmdItems, "-f")
	lstOfCmdItems = append(lstOfCmdItems, m_ScriptPath)
	return lstOfCmdItems
}
