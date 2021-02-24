package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
)

const (
	sharePath   = "/.local/share"
	filePath    = "clipGo"
	fileName    = "clipGo.json"
	clipCommand = "xsel"
)

type clipEntry struct {
}

var args = []string{"--output", "--clipboard"}

func main() {

	file := getFile()

	clipContent := getClipboardContent()

	file.Write(clipContent)
	file.Close()
}

func getFile() *os.File {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Error getting user's home dir: ", err)
	}

	userPath := usr.HomeDir

	os.Chdir(userPath + sharePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.MkdirAll(filePath, 0754)
	}

	os.Chdir(userPath + sharePath + "/" + filePath)

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0754)
	if err != nil {
		log.Fatal("Error opening the file: ", err)
	}

	return file
}

func getClipboardContent() []byte {
	clipContent, err := exec.Command(clipCommand, args[:]...).Output()
	if err != nil {
		log.Fatal("Error getting the content of clipboard: ", err)
	}
	return clipContent
}
