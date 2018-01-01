package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	jump "github.com/faceair/youjumpijump"
)

var train *jump.Train

func screenshot(filename string) image.Image {
	_, err := exec.Command("/system/bin/screencap", "-p", filename).Output()
	if err != nil {
		panic("screenshot failed")
	}

	inFile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	src, err := png.Decode(inFile)
	if err != nil {
		panic(err)
	}
	inFile.Close()
	return src
}

func main() {
	defer func() {
		jump.Debugger()
		if e := recover(); e != nil {
			log.Printf("%s: %s", e, debug.Stack())
			fmt.Print("the program has crashed, press any key to exit")
			var c string
			fmt.Scanln(&c)
		}
	}()

	var inputRatio float64
	var err error
	if len(os.Args) > 1 {
		inputRatio, err = strconv.ParseFloat(os.Args[1], 10)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Print("input jump ratio (recommend 2.04):")
		_, err = fmt.Scanln(&inputRatio)
		if err != nil {
			log.Fatal(err)
		}
	}

	out, err := exec.Command("/system/bin/sh", "/system/bin/wm", "size").Output()
	if err != nil {
		panic("wm size failed")
	}

	log.Printf("Get %s", out)
	size := regexp.MustCompile(`(\d+)x(\d+)`).FindStringSubmatch(string(out))
	width, err := strconv.Atoi(size[1])
	if err != nil {
		panic("wm size failed")
	}
	height, err := strconv.Atoi(size[2])
	if err != nil {
		panic("wm size failed")
	}
	train = jump.NewTrain(width, height, inputRatio, "Android")

	for {
		jump.Debugger()

		src := screenshot("jump.png")

		start, end := jump.Find(src)
		if start == nil {
			log.Print("can't find the starting point，please export the debugger directory")
			break
		} else if end == nil {
			log.Print("can't find the end point，please export the debugger directory")
			break
		}

		scale := float64(src.Bounds().Max.X) / 720
		nowDistance := jump.Distance(start, end)
		similarDistance, nowRatio := train.Find(nowDistance)

		log.Printf("from:%v to:%v distance:%.2f similar:%.2f ratio:%v press:%.2fms ", start, end, nowDistance, similarDistance, nowRatio, nowDistance*nowRatio)

		_, err = exec.Command("/system/bin/sh", "/system/bin/input", "swipe", strconv.FormatFloat(float64(start[0])*scale, 'f', 0, 32), strconv.FormatFloat(float64(start[1])*scale, 'f', 0, 32), strconv.FormatFloat(float64(end[0])*scale, 'f', 0, 32), strconv.FormatFloat(float64(end[1])*scale, 'f', 0, 32), strconv.Itoa(int(nowDistance*nowRatio))).Output()
		if err != nil {
			panic("touch failed")
		}

		go func() {
			time.Sleep(time.Millisecond * 170)
			src := screenshot("jump.test.png")

			finally, _ := jump.Find(src)
			if finally != nil {
				finallyDistance := jump.Distance(start, finally)
				finallyRatio := (nowDistance * nowRatio) / finallyDistance

				if finallyRatio > nowRatio/4*3 && finallyRatio < nowRatio*4/3 {
					train.Add(finallyDistance, finallyRatio)
				}
			}
		}()

		time.Sleep(time.Millisecond * 1500)
	}
}
