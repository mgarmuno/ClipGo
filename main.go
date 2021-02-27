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
	actionDelete  = "delete"
	dmenu         = "dmenu"
	echo          = "echo"
)

type clipEntry struct {
	Text string
}

var (
	xselArgs  = []string{"--output", "--clipboard"}
	echoArgs  = []string{"-e"}
	dmenuArgs = []string{"-l"}
)

func main() {

	action := os.Args[1]

	switch action {
	case actionAdd:
		addClipContentToFile()
	case actionShow:
		fileContent := getFileContent()
		showFileContentDmenu(fileContent)
	case actionDelete:
	}
}

func addClipContentToFile() {
	file := getFile()
	defer file.Close()

	clipContent := getClipboardContent()
	text := formatText(clipContent)
	if text == "" {
		return
	}

	fileContent := getFileContent()
	clipEntryContent := clipEntry{Text: text}
	fileContent = removeEquals(clipEntryContent, fileContent)
	fileContent = append([]clipEntry{clipEntryContent}, fileContent...)
	fileContent = removeTail(fileContent)

	baJSON, err := json.Marshal(fileContent)
	if err != nil {
		log.Fatal("Error creating json entry: ", err)
		return
	}

	file.Truncate(0)
	file.Seek(0, 0)

	file.Write(baJSON)
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
	clipContent, err := exec.Command(clipCommand, xselArgs[:]...).Output()
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

func getFileContent() []clipEntry {
	file := readFile()
	if string(file) == "" {
		return []clipEntry{}
	}
	var clipEntryArray []clipEntry
	err := json.Unmarshal(file, &clipEntryArray)
	if err != nil {
		log.Fatal("Error unmarshalling file: ", err)
	}

	return clipEntryArray
}

func showFileContentDmenu(ce []clipEntry) {
	entries := []string{}
	for _, s := range ce {
		entries = append(entries, s.Text)
	}

	stringForDm := strings.Join(entries, "\\n")
	echoArgs = append(echoArgs, stringForDm)
	dmenuArgs = append(dmenuArgs, fmt.Sprint(len(entries)))
	c1 := exec.Command(echo, echoArgs...)
	c2 := exec.Command(dmenu, dmenuArgs...)

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

func removeEquals(newEntry clipEntry, entries []clipEntry) []clipEntry {
	for i, entry := range entries {
		if entry == newEntry {
			copy(entries[i:], entries[i+1:])
			entries[len(entries)-1] = clipEntry{}
			entries = entries[:len(entries)-1]
		}
	}
	return entries
}

func removeTail(entries []clipEntry) []clipEntry {
	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	return entries
}
