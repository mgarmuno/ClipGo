package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	actionAdd     = "add"
	actionShow    = "show"
	dmenu         = "dmenu"
)

type clipEntry struct {
	Text string
}

var args = []string{"--output", "--clipboard"}

func main() {

	action := os.Args[1]

	switch action {
	case actionAdd:
		addClipContentToFile()
	case actionShow:
		showFileContent()
	}
}

func addClipContentToFile() {
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
	defer file.Close()

	return file
}

func readFile() []byte {
	fullFilePath := getFileFullPath()
	file, err := ioutil.ReadFile(fullFilePath)
	if err != nil {
		log.Fatal("Error reading file: ", err)
		return []byte("")
	}

	return file
}

func getFileFullPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Error getting user's home dir: ", err)
	}
	userPath := usr.HomeDir
	return userPath + sharePath + "/" + filePath + "/" + fileName

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

func showFileContent() {
	file := readFile()
	var ce []clipEntry
	err := json.Unmarshal(file, &ce)
	if err != nil {
		log.Fatal("Error unmarshalling file: ", err)
	}

	showFileContentDmenu(ce)
}

func showFileContentDmenu(ce []clipEntry) {
	entries := []string{}
	for _, s := range ce {
		entries = append(entries, s.Text)
	}

	stringForDm := strings.Join(entries, "\\n")

	c1 := exec.Command("echo", "-e", stringForDm)
	c2 := exec.Command(dmenu, "-l", fmt.Sprint(len(entries)))

	r, w := io.Pipe()

	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()

	st := b2.String()

	fmt.Println(st)
}
