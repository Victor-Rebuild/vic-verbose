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

	"github.com/kercre123/vector-gobot/pkg/vbody"
	"github.com/kercre123/vector-gobot/pkg/vscreen"
)

// program which will run in recovery partition

var CurrentList *List
var ScreenInited bool
var BodyInited bool
var MaxTM uint32
var MinTM uint32
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

type OTA struct {
	Name string
	URL  string
}

func (c *List) MoveDown() {
	if c.Len == c.Position {
		c.Position = 1
	} else {
		c.Position = c.Position + 1
	}
	c.UpdateScreen()
}

func (c *List) MoveUp() {
	// i'm not sure how to determine direction from the encoders, so i am doing always down
	fmt.Println("up")
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

func StartAnki_Confirm() {
	c := *CurrentList
	CurrentList = Confirm_Create_Anki(StartAnki, c)
	CurrentList.Init()
}

func StartAnki() {
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

func StartRescue() {
	KillButtonDetect := false
	// rescue can crash, often
	HangBody = true
	scrnData := vscreen.CreateTextImage("vic-rescue will start in 3 seconds. Press the button anytime to return to the menu.")
	vscreen.SetScreen(scrnData)
	vscreen.StopLCD()
	ScreenInited = false
	time.Sleep(time.Second * 3)
	cmd := exec.Command("/bin/bash", "-c", "/anki/bin/vic-rescue")
	go func() {
		for {
			frame := GetFrame()
			if frame.ButtonState || KillButtonDetect {
				break
			}
			time.Sleep(time.Millisecond * 10)
		}
		fmt.Println("killing rescue")
		cmd.Process.Kill()
	}()
	cmd.Run()
	CurrentList = Recovery_Create()
	CurrentList.Init()
	time.Sleep(time.Second / 3)
	HangBody = false
}

func Reboot_Do() {
	exec.Command("/bin/bash", "-c", "bootctl f set_active a")
	scrnData := vscreen.CreateTextImage("Rebooting...")
	vscreen.SetScreen(scrnData)
	StopListening = true
	time.Sleep(time.Second / 2)
	vbody.StopSpine()
	vscreen.StopLCD()
	exec.Command("/bin/bash", "-c", "reboot").Run()
}

func Reboot_Create() *List {
	// "ARE YOU SURE?"
	var Reboot List

	Reboot.Info = "Reboot?"
	Reboot.InfoColor = color.RGBA{0, 255, 0, 255}
	Reboot.ClickFunc = []func(){Reboot_Do, func() {
		CurrentList = Recovery_Create()
		CurrentList.Init()
	}}

	Reboot.Lines = []vscreen.Line{
		{
			Text:  "Yes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "No",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Reboot
}

func ClearUserData_Do() {
	vscreen.SetScreen(vscreen.CreateTextImage("Clearing User Data..."))
	exec.Command("/bin/bash", "-c", "blkdiscard -s /dev/block/bootdevice/by-name/userdata").Run()
	exec.Command("/bin/bash", "-c", "blkdiscard -s /dev/block/bootdevice/by-name/switchboard").Run()
	Reboot_Do()
}

func ClearUserData_Create() *List {
	// "ARE YOU SURE?"
	var Reboot List

	Reboot.Info = "Clear user data?"
	Reboot.InfoColor = color.RGBA{0, 255, 0, 255}
	Reboot.ClickFunc = []func(){ClearUserData_Do, func() {
		CurrentList = Recovery_Create()
		CurrentList.Init()
	}}

	Reboot.Lines = []vscreen.Line{
		{
			Text:  "Yes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "No",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Reboot
}

func DetectButtonPress() {
	// for functions which show on screen, but aren't lists. hangs ListenToBody, returns when button is presed
	for {
		frame := GetFrame()
		if frame.ButtonState {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}

}

func Recovery_Create() *List {
	var Test List

	Test.Info = "impl"
	Test.InfoColor = color.RGBA{0, 255, 0, 255}

	Test.ClickFunc = []func(){
		StartAnki_Confirm,
	}

	Test.Lines = []vscreen.Line{
		{
			Text:  "Watch logs",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Test
}

func Confirm_Create_Anki(do func(), origList List) *List {
	// "ARE YOU SURE?"
	var Test List

	Test.Info = "See logs?"
	Test.InfoColor = color.RGBA{0, 255, 0, 255}
	Test.ClickFunc = []func(){do, func() {
		CurrentList = &origList
		CurrentList.Init()
	}}

	Test.Lines = []vscreen.Line{
		{
			Text:  "Yes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "No",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Test
}

func Confirm_Install_OTA(do func(), origList List) *List {
	// "ARE YOU SURE?"
	var Test List

	Test.Info = "Install this OTA?"
	Test.InfoColor = color.RGBA{0, 255, 0, 255}
	Test.ClickFunc = []func(){do, func() {
		CurrentList = &origList
		CurrentList.Init()
	}}

	Test.Lines = []vscreen.Line{
		{
			Text:  "Yes",
			Color: color.RGBA{255, 255, 255, 255},
		},
		{
			Text:  "No",
			Color: color.RGBA{255, 255, 255, 255},
		},
	}

	return &Test
}

func TestIfBodyWorking() {
	// if body isn't working, start anki processes
	err := vbody.InitSpine()
	if err != nil {
		vscreen.InitLCD()
		vscreen.BlackOut()
		data := vscreen.CreateTextImage("Error! Not able to communicate with the body. Starting Anki processes...")
		vscreen.SetScreen(data)
		vbody.StopSpine()
		vscreen.StopLCD()
		exec.Command("/bin/bash", "-c", "systemctl start anki-robot.target").Run()
		os.Exit(0)
	} else {
		BodyInited = true
	}
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

	StartAnki()
	CurrentList.Init()
	vbody.SetLEDs(vbody.LED_OFF, vbody.LED_OFF, vbody.LED_OFF)
	fmt.Println("started")
	InitFrameGetter()
	ListenToBody()
}
