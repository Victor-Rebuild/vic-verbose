package main

import (
	"bufio"
	"fmt"
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

func (c *List) MoveDown() {
	if c.Len == c.Position {
		c.Position = 1
	} else {
		c.Position = c.Position + 1
	}
	c.UpdateScreen()
}

func (c *List) UpdateScreen() {
	var linesShow []vscreen.Line
	// if info, have list go to bottom
	// 7 lines fit comfortably on screen
	if c.Info != "" {
		newLine := vscreen.Line{
			Text:  c.Info,
			Color: c.InfoColor,
		}
		linesShow = append(linesShow, newLine)
		numOfSpaces := 7 - c.Len
		if numOfSpaces < 0 {
			panic("too many items in list" + fmt.Sprint(numOfSpaces))
		}
		for i := 2; i < numOfSpaces; i++ {
			newLine = vscreen.Line{
				Text:  " ",
				Color: c.InfoColor,
			}
			linesShow = append(linesShow, newLine)
		}
	}
	for i, line := range c.Lines {
		var newLine vscreen.Line
		if i == c.Position-1 {
			newLine.Text = "> " + line.Text
			newLine.Color = line.Color
		} else {
			newLine.Text = "  " + line.Text
			newLine.Color = line.Color
		}
		linesShow = append(linesShow, newLine)
	}
	scrnData := vscreen.CreateTextImageFromLines(linesShow)
	vscreen.SetScreen(scrnData)
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
	c.UpdateScreen()
	c.inited = true
}

func ListenToBody() {
	if !CurrentList.inited {
		fmt.Println("error: init list before listening dummy")
		os.Exit(1)
	}
	for {
		if StopListening {
			fmt.Println("not listening anymore")
			StopListening = false
			return
		}
		if !CurrentList.inited || HangBody {
			for {
				time.Sleep(time.Second / 5)
				if CurrentList.inited && !HangBody {
					break
				}
			}
		}
		frame := GetFrame()
		if frame.ButtonState {
			CurrentList.ClickFunc[CurrentList.Position-1]()
			time.Sleep(time.Second / 3)
		}
		for i, enc := range frame.Encoders {
			if i > 1 {
				// only read wheels
				break
			}
			if enc.DLT < -1 {
				stopTimer := false
				stopWatch := false
				go func() {
					timer := 0
					for {
						if StopListening {
							fmt.Println("not listening anymore")
							StopListening = false
							return
						}
						if stopTimer {
							break
						}
						if timer == 30 {
							CurrentList.MoveDown()
							stopWatch = true
							break
						}
						timer = timer + 1
						time.Sleep(time.Millisecond * 10)
					}
				}()
				for {
					if StopListening {
						fmt.Println("not listening anymore")
						StopListening = false
						return
					}
					frame = GetFrame()
					if stopWatch {
						break
					}
					if frame.Encoders[i].DLT == 0 {
						stopTimer = true
						break
					}
				}
			}
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func StartLogging() {
	// scrnData := vscreen.CreateTextImage("To come back to this menu, go to CCIS and select `MENU` or `BACK TO MENU`. Starting in 3 seconds...")
	// vscreen.SetScreen(scrnData)
	// time.Sleep(time.Second * 4)
	// scrnData = vscreen.CreateTextImage("Stopping body...")
	// vscreen.SetScreen(scrnData)
	// CurrentList.inited = false
	// time.Sleep(time.Second / 3)
	// StopFrameGetter()
	// vbody.StopSpine()
	// scrnData := vscreen.CreateTextImage("Grabbing logs")
	// vscreen.SetScreen(scrnData)
	// vscreen.StopLCD()
	// ScreenInited = false
	// BodyInited = false
	time.Sleep(time.Second / 2)
	// exec.Command("/bin/bash", "-c", "systemctl start anki-robot.target").Run()
	// watch logcat for clear user data screen // journalctl since we don't have logcat
	cmd := exec.Command("journalctl", "-f", "-n", "10")
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

func Recovery_Create() *List {
	var Test List

	Test.Info = "If you see this, logging has failed"
	Test.InfoColor = color.RGBA{0, 255, 0, 255}
	return &Test
}

func main() {
	vbody.InitSpine()
	BodyInited = true

	vbody.SetLEDs(vbody.LED_OFF, vbody.LED_OFF, vbody.LED_RED)
	time.Sleep(time.Second)
	vbody.SetLEDs(vbody.LED_OFF, vbody.LED_GREEN, vbody.LED_RED)
	time.Sleep(time.Second)
	vbody.SetLEDs(vbody.LED_BLUE, vbody.LED_GREEN, vbody.LED_RED)
	CurrentList = Recovery_Create()

	vscreen.InitLCD()
	vscreen.BlackOut()
	ScreenInited = true

	RandomLights()

	StartLogging()
	CurrentList.Init()
}
