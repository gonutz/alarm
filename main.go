package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gonutz/w32"
	"github.com/gonutz/win"
)

var (
	atTime = flag.String("at", "", "time at which to ring the alarm in the form "+time.Kitchen)
	inTime = flag.String("in", "", "time after which to ring the alarm from now")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\nExactly one of these arguments must be valid\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// exactly one of the arguments has to be non-empty
	if (*atTime == "") == (*inTime == "") {
		flag.Usage()
		return
	}

	var waitTime time.Duration
	if *atTime != "" {
		at, err := time.Parse(time.Kitchen, *atTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing time at which to ring the alarm:", err)
			flag.Usage()
			return
		}
		now := time.Now()
		// at has only the hour and minutes set, everything else is 0 so make it
		// the same day or one day after (if the time has already passed today)
		at = now.
			Add(time.Duration(at.Hour()-now.Hour()) * time.Hour).
			Add(time.Duration(at.Minute()-now.Minute()) * time.Minute).
			Add(-time.Duration(now.Second()) * time.Second).
			Add(-time.Duration(now.Nanosecond()) * time.Nanosecond)
		if !at.After(now) {
			at = at.Add(24 * time.Hour)
		}
		waitTime = at.Sub(now)
	} else {
		in, err := time.ParseDuration(*inTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing time at which to ring the alarm:", err)
			flag.Usage()
			return
		}
		waitTime = in
	}

	win.HideConsoleWindow()
	time.Sleep(waitTime)
	w32.MessageBeep(0)
	window, err := win.NewWindow(
		0, 0, 600, 600, "alarm_window",
		func(window w32.HWND, msg uint32, w, l uintptr) uintptr {
			if msg == w32.WM_DESTROY {
				w32.PostQuitMessage(0)
				return 0
			} else {
				return w32.DefWindowProc(window, msg, w, l)
			}
		})
	w32.SetWindowText(window, "Alarm")
	if err != nil {
		panic(err)
	}
	win.RunMainLoop()
}
