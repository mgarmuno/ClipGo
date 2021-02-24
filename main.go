package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

const (
	relocateFirst = true
	maxEntries    = 10
	sharePath     = "/.local/share"
	filePath      = "clipGo"
	fileName      = "clipGo.json"
	clipCommand   = "xsel"
)

type clipEntry struct {
	Text string
}

var args = []string{"--output", "--clipboard"}

func main() {

	file := getFile()

	clipContent := getClipboardContent()
	text := formatText(clipContent)
	ce := clipEntry{Text: text}

	ca := make([]clipEntry, 2)

	ca[0] = ce
	ca[1] = ce

	json, err := json.Marshal(ca)
	if err != nil {
		log.Fatal("Error creating json entry: ", err)
	}

	file.Truncate(0)
	file.Seek(0, 0)

	file.Write(json)
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
		os.MkdirAll(filePath, 0700)
	}

	os.Chdir(userPath + sharePath + "/" + filePath)

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0700)
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

func formatText(clipContent []byte) string {
	var lanes []string

	reader := bytes.NewReader(clipContent)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lanes = append(lanes, scanner.Text())
	}

	text := strings.Join(lanes, "\\n")

	return text
}
