package main

import (
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
	if len(os.Args) < 2 {
		log.Fatal("At least one argument")
	}
	action := os.Args[1]

	switch action {
	case actionAdd:
		text := getClipboardContent()
		addTextToFile(text)
	case actionShow:
		showEntities()
	case actionDelete:
		deleteEntity()
	default:
		fmt.Println("Please, use one of the following valid arguments:")
		fmt.Println("add    -> Adds a new entry to the history")
		fmt.Println("show   -> Shows the history")
		fmt.Println("delete -> Shows the history and deletes the select entry")
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

func showEntities() {
	entities := getFileContent()
	selText := showEntitiesDmenu(entities)

	if isValidForSave(selText) {
		setSelectedItem(selText, entities)
	}
}

func deleteEntity() {
	entities := getFileContent()
	selText := showEntitiesDmenu(entities)
	index := strings.Split(selText, dmenuSeparator)[0]
	i, err := strconv.Atoi(index)
	if err != nil {
		log.Fatal("Error parsing string index to int on delete: ", err)
	}
	entities = removeEntityByIndex(i, entities)
	marshalAndSave(entities)
}

func showEntitiesDmenu(entities []entity) string {
	entries := []string{}
	for _, ent := range entities {
		cleanedUpText := cleanTextForDmenu(ent.Text)
		entries = append(entries, fmt.Sprint(ent.Position)+dmenuSeparator+cleanedUpText)
	}

	stringForDm := strings.Join(entries, "\\n")
	return executeCommands(stringForDm, len(entries))
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

func executeCommands(ent string, len int) string {
	echoArgs = append(echoArgs, ent)
	dmenuArgs = append(dmenuArgs, fmt.Sprint(len))
	cmdEcho := exec.Command(echo, echoArgs...)
	cmdDmenu := exec.Command(dmenu, dmenuArgs...)

	read, write := io.Pipe()

	cmdEcho.Stdout = write
	cmdDmenu.Stdin = read

	var output bytes.Buffer
	cmdDmenu.Stdout = &output

	cmdEcho.Start()
	cmdDmenu.Start()
	cmdEcho.Wait()
	write.Close()
	cmdDmenu.Wait()

	return output.String()
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
		log.Panic("Error generatin\nd cmd to write in clipboard: ", err)
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
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, "\n", endLineSign)
	s = strings.ReplaceAll(s, "\t", "    ")

	return s
}

func removeEquals(newEntry entity, entries []entity) []entity {
	newEntities := []entity{}
	for _, entry := range entries {
		if entry.Text == newEntry.Text {
			continue
		}
		newEntities = append(newEntities, entry)
	}

	return newEntities
}

func removeTail(entries []entity) []entity {
	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	return entries
}

func removeEntityByIndex(i int, entities []entity) []entity {
	copy(entities[i:], entities[i+1:])
	entities[len(entities)-1] = entity{}
	return entities[:len(entities)-1]
}
