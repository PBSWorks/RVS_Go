package common

import (
	"log"
	"os"
	"path/filepath"
)

func AllocateUniqueFolder(ParentsDir string, prefix string) string {
	var folderpath string = filepath.Join(ParentsDir, getRandomString(prefix))
	if _, err := os.Stat(folderpath); os.IsNotExist(err) {
		os.MkdirAll(folderpath, os.ModePerm)
	}
	return folderpath
}

func AllocateFile(sFileName string, sParentDirAbsPath string, username string, password string) string {

	if !IsValidString(sFileName) {
		sFileName = TEMP_FILE_NAME
	}
	var file string = filepath.Join(sParentDirAbsPath, sFileName)
	arrCmd := []string{}
	scriptPath, err := filepath.Abs(getCreateFileScriptPath())
	if err != nil {
		log.Fatal(err)
	}
	arrCmd = append(arrCmd, scriptPath)
	arrCmd = append(arrCmd, "FILE")

	if IsWindows() {
		arrCmd = append(arrCmd, file)
	} else {
		arrCmd = append(arrCmd, file)
	}
	RunCommand(arrCmd, username, password)
	return file
}

func getCreateFileScriptPath() string {
	var path string = ""
	if IsWindows() {
		if Is32BitOS() {
			path = GetRSHome() + "/bin/win32/CreateFile.bat"
		} else {
			path = GetRSHome() + "\\bin\\win64\\CreateFile.bat"
		}
	} else {
		if Is32BitOS() {
			path = GetRSHome() + "/bin/linux32/CreateFile.sh"
		} else {
			path = GetRSHome() + "/bin/linux64/CreateFile.sh"
		}
	}
	return path
}

func AllocateFileWithGlobalPermission(sFileName string, sParentDirAbsPath string) string {
	log.Println("Entering method allocateFileWithGlobalPermission")
	log.Println("Creating file " + sParentDirAbsPath + "/" + sFileName + "with global permission for every one")
	myfile, e := os.Create(sParentDirAbsPath + "/" + sFileName)
	if e != nil {
		log.Fatal(e)
	}
	log.Println(myfile)
	myfile.Chmod(0777)
	myfile.Close()
	log.Println("Created file " + sParentDirAbsPath + "/" + sFileName + " with global permission for every one")
	return sParentDirAbsPath + "/" + sFileName
}
