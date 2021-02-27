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
	"strconv"
	"strings"
)

const (
	relocateFirst  = true
	maxEntries     = 10
	sharePath      = "/.local/share"
	filePath       = "clipGo"
	fileName       = "clipGo.json"
	clipCommand    = "xsel"
	actionAdd      = "add"
	actionShow     = "show"
	actionDelete   = "delete"
	dmenu          = "dmenu"
	echo           = "echo"
	dmenuSeparator = " => "
	endLineSign    = "⏎"
)

type entity struct {
	Position int
	Text     string
}

var (
	xselOutArgs = []string{"--output", "--clipboard"}
	xselInArgs  = []string{"--input", "--clipboard"}
	echoArgs    = []string{"-e"}
	dmenuArgs   = []string{"-l"}
)

func main() {
	action := os.Args[1]
	switch action {
	case actionAdd:
		clipContent := getClipboardContent()
		addTextToFile(clipContent)
	case actionShow:
		fileContent := getFileContent()
		showFileContentDmenu(fileContent)
	case actionDelete:
	}
}

func addTextToFile(text string) {
	if !isValidForSave(text) {
		return
	}

	clipEntryContent := entity{Text: text}

	fileContent := getFileContent()
	fileContent = removeEquals(clipEntryContent, fileContent)
	fileContent = append([]entity{clipEntryContent}, fileContent...)
	fileContent = removeTail(fileContent)

	marshalAndSave(fileContent)
}

func isValidForSave(text string) bool {
	switch {
	case text == "":
		return false
	case strings.ReplaceAll(text, "\t", "") == "":
		return false
	}

	return true
}

func marshalAndSave(entries []entity) {
	entries = assignOrderNumbers(entries)

	baJSON, err := json.Marshal(entries)
	if err != nil {
		return
	}

	writeJSONOnFile(baJSON)
}

func writeJSONOnFile(baJSON []byte) {
	file := getFile()

	file.Truncate(0)
	file.Seek(0, 0)

	file.Write(baJSON)
	file.Close()
}

func assignOrderNumbers(entries []entity) []entity {
	for i := range entries {
		entries[i].Position = i
	}

	return entries
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

func getClipboardContent() string {
	clipContent, err := exec.Command(clipCommand, xselOutArgs[:]...).Output()
	if err != nil {
		log.Fatal("Error getting the content of clipboard: ", err)
	}

	return string(clipContent)
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

func getFileContent() []entity {
	file := readFile()
	if string(file) == "" {
		return []entity{}
	}
	var clipEntryArray []entity
	err := json.Unmarshal(file, &clipEntryArray)
	if err != nil {
		log.Fatal("Error unmarshalling file: ", err)
	}

	return clipEntryArray
}

func showFileContentDmenu(fileEnt []entity) {
	entries := []string{}
	for _, ent := range fileEnt {
		cleanedUpText := cleanTextForDmenu(ent.Text)
		entries = append(entries, fmt.Sprint(ent.Position)+dmenuSeparator+cleanedUpText)
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

	selText := b2.String()

	if isValidForSave(selText) {
		setSelectedItem(selText, fileEnt)
	}
}

func setSelectedItem(selText string, entries []entity) {
	if selText == "" {
		return
	}
	as := strings.Split(selText, dmenuSeparator)
	position, err := strconv.Atoi(as[0])

	if err != nil {
		log.Panic("Error selecting item: ", err)
	}

	if relocateFirst {
		addTextToFile(entries[position].Text)
	}

	writeToClipboard(entries[position].Text)
}

func writeToClipboard(st string) {
	cmd := exec.Command(clipCommand, xselInArgs[:]...)
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Panic("Error generatind cmd to write in clipboard: ", err)
	}

	if err := cmd.Start(); err != nil {
		log.Panic("Error starting cmd: ", err)
	}

	if _, err := in.Write([]byte(st)); err != nil {
		log.Panic("Error writing in clipboard: ", err)
	}

	if err := in.Close(); err != nil {
		log.Panic("Error closing in pipe: ", err)
	}

	cmd.Wait()
}

func cleanTextForDmenu(s string) string {
	s = strings.ReplaceAll(s, "\n", endLineSign)
	s = strings.ReplaceAll(s, "\t", "    ")

	return s
}

func removeEquals(newEntry entity, entries []entity) []entity {
	for i, entry := range entries {
		if entry.Text == newEntry.Text {
			copy(entries[i:], entries[i+1:])
			entries[len(entries)-1] = entity{}
			entries = entries[:len(entries)-1]
		}
	}

	return entries
}

func removeTail(entries []entity) []entity {
	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	return entries
}
