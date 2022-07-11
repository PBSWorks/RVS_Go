package common

import (
	"altair/rvs/utils"
	"log"
	"os"
	"path/filepath"

	l "altair/rvs/globlog"
)

func AllocateUniqueFolder(ParentsDir string, prefix string) string {
	var folderpath string = filepath.Join(ParentsDir, utils.GetRandomString(prefix))
	if _, err := os.Stat(folderpath); os.IsNotExist(err) {
		os.MkdirAll(folderpath, os.ModePerm)
	}
	return folderpath
}

func AllocateFile(sFileName string, sParentDirAbsPath string, username string, password string) string {

	if !utils.IsValidString(sFileName) {
		sFileName = utils.TEMP_FILE_NAME
	}
	var file string = filepath.Join(sParentDirAbsPath, sFileName)
	arrCmd := []string{}
	scriptPath, err := filepath.Abs(getCreateFileScriptPath())
	if err != nil {
		log.Fatal(err)
	}
	arrCmd = append(arrCmd, scriptPath)
	arrCmd = append(arrCmd, "FILE")

	if utils.IsWindows() {
		arrCmd = append(arrCmd, file)
	} else {
		arrCmd = append(arrCmd, file)
	}
	RunCommand(arrCmd, username, password)
	return file
}

func getCreateFileScriptPath() string {
	var path string = ""
	if utils.IsWindows() {
		if utils.Is32BitOS() {
			path = utils.GetRSHome() + "/bin/win32/CreateFile.bat"
		} else {
			path = utils.GetRSHome() + "\\bin\\win64\\CreateFile.bat"
		}
	} else {
		if utils.Is32BitOS() {
			path = utils.GetRSHome() + "/bin/linux32/CreateFile.sh"
		} else {
			path = utils.GetRSHome() + "/bin/linux64/CreateFile.sh"
		}
	}
	return path
}

func AllocateFileWithGlobalPermission(sFileName string, sParentDirAbsPath string) string {
	l.Log().Info("Entering method allocateFileWithGlobalPermission")
	l.Log().Info("Creating file " + sParentDirAbsPath + "/" + sFileName + "with global permission for every one")
	myfile, e := os.Create(sParentDirAbsPath + "/" + sFileName)
	if e != nil {
		log.Fatal(e)
	}
	l.Log().Info(myfile)
	myfile.Chmod(0777)
	myfile.Close()
	l.Log().Info("Created file " + sParentDirAbsPath + "/" + sFileName + " with global permission for every one")
	return sParentDirAbsPath + "/" + sFileName
}
