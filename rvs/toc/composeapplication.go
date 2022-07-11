package toc

import (
	"altair/rvs/common"
	l "altair/rvs/globlog"
	"altair/rvs/utils"
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
	l.Log().Info("Compose Execution Time: ", diff)
}

func buildCommandArrayForOperatingSystem() []string {
	lstOfCmdItems := []string{}
	var sExec string
	if utils.IsWindows() {
		sExec = common.GetProductInstallationLocation(utils.COMPOSE_PRODUCT_ID) + utils.COMPOSE_WINDOWS_64BIT_EXEC
		lstOfCmdItems = append(lstOfCmdItems, utils.GetPlatformIndependentFilePath(sExec, true))

	} else {
		sExec = common.GetProductInstallationLocation(utils.COMPOSE_PRODUCT_ID) + utils.COMPOSE_UNIX_EXEC
		lstOfCmdItems = append(lstOfCmdItems, utils.GetPlatformIndependentFilePath(sExec, true))
	}
	info, err := os.Stat(sExec)
	if os.IsNotExist(err) {
		l.Log().Info(info)
		l.Log().Info("sExec: ", sExec)
		l.Log().Info("Compose Execution file does not exists.")
	}
	lstOfCmdItems = append(lstOfCmdItems, "-f")
	lstOfCmdItems = append(lstOfCmdItems, m_ScriptPath)
	return lstOfCmdItems
}
