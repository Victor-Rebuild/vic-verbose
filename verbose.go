package main

import (
	"bufio"
	"image/color"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/os-vector/vector-gobot/pkg/vbody"
	"github.com/os-vector/vector-gobot/pkg/vscreen"
)

// program which will display logs on boot

var CurrentList *List
var ScreenInited bool
var BodyInited bool
var StopListening bool
var HangBody bool

type List struct {
	Info      string
	InfoColor color.Color
	Lines     []vscreen.Line
	// len and position start with 1
	Len       int
	Position  int
	ClickFunc []func()
	inited    bool
}

func (c *List) Init() {
	c.Position = 1
	c.Len = len(c.Lines)
	if !BodyInited {
		vbody.InitSpine()
		InitFrameGetter()
		BodyInited = true
	}
	if !ScreenInited {
		vscreen.InitLCD()
		vscreen.BlackOut()
		ScreenInited = true
	}
	c.inited = true
}

func StartLogging() {
	cmd := exec.Command("journalctl", "-f", "-n", "1")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		scrnData := vscreen.CreateTextImage(line)

		vscreen.SetScreen(scrnData)

		RandomLights()
		if strings.Contains(line, "Starting Victor init") {
			break
		}
	}

	scrnData := vscreen.CreateTextImage("Starting Vector processes")
	vscreen.SetScreen(scrnData)

	vbody.StopSpine()
	vscreen.StopLCD()
	ScreenInited = false
	BodyInited = false
	os.Exit(0)
}

func RandomLights() {
	colors := []uint32{vbody.LED_BLUE, vbody.LED_GREEN, vbody.LED_RED, vbody.LED_OFF}

	color1 := colors[rand.Intn(len(colors))]
	color2 := colors[rand.Intn(len(colors))]
	color3 := colors[rand.Intn(len(colors))]

	vbody.SetLEDs(color1, color2, color3)
}

func Failed() *List {
	var Test List

	Test.Info = "If you see this, logging has failed"
	Test.InfoColor = color.RGBA{255, 0, 0, 255}
	return &Test
}

func main() {
	vbody.InitSpine()
	BodyInited = true

	RandomLights()
	time.Sleep(time.Second / 2)
	RandomLights()
	time.Sleep(time.Second / 2)
	RandomLights()
	CurrentList = Failed()
	RandomLights()
	vscreen.InitLCD()
	RandomLights()
	vscreen.BlackOut()
	RandomLights()
	ScreenInited = true

	RandomLights()

	StartLogging()
	CurrentList.Init()
}
