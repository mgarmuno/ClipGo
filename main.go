package main

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	buf := getImageBytes()

	systray.SetIcon(buf.Bytes())
	systray.SetTooltip("A ver si podemos hacer esta mierda")
}

func onExit() {

}

func getImageBytes() *bytes.Buffer {
	f, err := os.Open("icons/icon-white.png")
	if err != nil {
		log.Panic("Error reading icon file: ", err)
	}

	image, _, err := image.Decode(f)
	if err != nil {
		log.Panic("Error decoding icon: ", err)
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, image)
	if err != nil {
		log.Panic("Error encoding icon: ", err)
	}

	return buf
}
