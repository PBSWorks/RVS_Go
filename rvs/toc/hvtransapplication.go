package toc

import (
	"altair/rvs/common"
	"fmt"
	"os"
	"time"
)

var HVTRANS_WINDOWS32_EXEC = "\\io\\result_readers\\bin\\win32\\hvtrans.exe"
var HVTRANS_WINDOWS64_EXEC = "\\io\\result_readers\\bin\\win64\\hvtrans.exe"
var HVTRANS_UNIX_EXEC = "/scripts/hvtrans"

func executeAnimationApplication(sConfigFilePath string, sResultFilePath string, username string, password string) {
	//common.Readconfigfile()
	startdt := time.Now()
	common.RunCommand(buildCommandArray(sConfigFilePath, sResultFilePath, sResultFilePath), username, password)
	enddt := time.Now()
	diff := enddt.Sub(startdt)
	fmt.Println(diff)
}

func buildCommandArray(sConfigFilePath string, sResultFilePath string, sModelFilePath string) []string {
	lstOfCmdItems := []string{}
	var sExec string
	if common.IsWindows() {
		if common.Is32BitOS() {
			sExec = common.GetProductInstallationLocation(common.HYPERWORKS_PRODUCT_ID) + HVTRANS_WINDOWS32_EXEC
		} else {
			sExec = common.GetProductInstallationLocation(common.HYPERWORKS_PRODUCT_ID) + HVTRANS_WINDOWS64_EXEC
		}
		/* Do not get platform independent path, Since, For AIF Impersonation the HVTrans
		 * path should contain \ backslash (Like: C:\Altair\hw10.0\io\...) */
		//	sExec = sExec.replace("/", "\\")
		lstOfCmdItems = append(lstOfCmdItems, sExec)
	} else {
		if common.Is32BitOS() {
			sExec = common.GetProductInstallationLocation(common.HYPERWORKS_PRODUCT_ID) + HVTRANS_UNIX_EXEC
		} else {
			sExec = common.GetProductInstallationLocation(common.HYPERWORKS_PRODUCT_ID) + HVTRANS_UNIX_EXEC
		}
		lstOfCmdItems = append(lstOfCmdItems, common.GetPlatformIndependentFilePath(sExec, true))
		lstOfCmdItems = append(lstOfCmdItems, "-nobg")
	}

	info, err := os.Stat(sExec)
	if os.IsNotExist(err) {
		fmt.Println(info)
		fmt.Println("HVtrans Execution file does not exists.")
	}
	configfileInfo, configerr := os.Stat(sConfigFilePath)
	if os.IsNotExist(configerr) {
		fmt.Println(configfileInfo)
		fmt.Println("Config script path not found.")
	}

	// No space between -c and config file path as per hvtrans documentation
	if common.IsValidString(sConfigFilePath) {
		lstOfCmdItems = append(lstOfCmdItems, "-c"+common.GetPlatformIndependentFilePath(sConfigFilePath, true))
	}
	// Result file
	lstOfCmdItems = append(lstOfCmdItems, common.GetPlatformIndependentFilePath(sResultFilePath, true))
	//Model file path
	lstOfCmdItems = append(lstOfCmdItems, common.GetPlatformIndependentFilePath(sModelFilePath, true))

	// Compression percentage
	// if (IsValidString(sCompression) )
	// {
	//     lstOfCmdItems.add("-z" + sCompression);
	// }
	fmt.Println(lstOfCmdItems)

	return lstOfCmdItems

}
