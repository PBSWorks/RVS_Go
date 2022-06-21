package common

import (
	"fmt"
	"os/exec"
	"time"
)

var isAIFImpersonationEnabled bool = true

const SPACE_CHAR = " "
const ALTAIR_LICENSE_PATH_VALUE = "6200@10.145.13.38"

func RunCommand(sArrCmd []string, username string, password string) int {
	if isAIFImpersonationEnabled {
		//	fmt.Println("Running with impersonation on")
		return runCommandWithAIFImpersonation(sArrCmd[:], username, password, false)
	} else {
		fmt.Println("Running with impersonation off")
		return -1
	}
}

func runCommandWithAIFImpersonation(sArrCmd []string, username string, password string, bNeedXvfb bool) int {

	if len(sArrCmd) <= 0 {
		fmt.Println("Recieved command is empty")
		return -1
	}

	// Initialize Environement Parameters for AIF Impersonation
	sArrEnvironment := []string{}
	for i := 0; i < len(sArrCmd); i++ {
		sArrEnvironment = append(sArrEnvironment, sArrCmd[i])
	}
	//sArrEnvironment = append(sArrEnvironment, os.Environ()...)

	//sArrEnvironment = append(sArrEnvironment, "AIF_IMPERSONATION_USER="+username)
	//sArrEnvironment = append(sArrEnvironment, "AIF_IMPERSONATION_PASSWORD="+password)
	//sArrEnvironment = append(sArrEnvironment, "AIF_IMPERSONATION_WINSTA="+getRandomString())
	//sArrEnvironment = append(sArrEnvironment, "XVFB_DISPLAY="+"11")
	//sArrEnvironment = append(sArrEnvironment, "ALTAIR_LICENSE_PATH="+getLicensePath())

	var runnerexecpath = ""
	if IsWindows() {
		if Is32BitOS() {
			runnerexecpath = GetRSHome() + "/bin/win32/ProcessRunner.exe"
		} else {
			runnerexecpath = GetRSHome() + "/bin/win64/ProcessRunner.exe"
		}
	} else {
		if bNeedXvfb {
			// if (!m_bXvfbStarted) {
			// 	this.startXvfb(m_rmFramework.getRMSiteConfiguration().getXvfbDisplay());
			// 	m_bXvfbStarted = true;
			// }
			if Is32BitOS() {
				runnerexecpath = GetRSHome() + "/bin/linux32/ImpersonatedProcessRunner_Xvfb.sh"
			} else {
				runnerexecpath = GetRSHome() + "/bin/linux64/ImpersonatedProcessRunner_Xvfb.sh"
			}
		} else {
			if Is32BitOS() {
				runnerexecpath = GetRSHome() + "/bin/linux32/ImpersonatedProcessRunner.sh"
			} else {
				runnerexecpath = GetRSHome() + "/bin/linux64/ImpersonatedProcessRunner.sh"
			}
		}
	}

	fmt.Println("Command:", sArrEnvironment)
	dtstart := time.Now()

	cmd := exec.Command(runnerexecpath, sArrEnvironment...)
	var exitCode = 1
	// Does not wait for command to complete before returning
	if err := cmd.Start(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Exit code is %d\n", exitError.ExitCode())
			exitCode = 1
		}
	}

	// Wait for cmd to Return
	if err := cmd.Wait(); err != nil {

		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Exit code is %d\n", exitError.ExitCode())
			exitCode = 1
		}
	}

	dtend := time.Now()
	diff := dtend.Sub(dtstart)
	fmt.Println("Actual Command Execution Time: ", diff)

	return exitCode
}
