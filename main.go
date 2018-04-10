package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gonutz/w32"
	"github.com/gonutz/win"
)

var (
	atTime = flag.String("at", "", "time at which to ring the alarm in the form "+time.Kitchen)
	inTime = flag.String("in", "", "time after which to ring the alarm from now, e.g. 5s or 1h15m")
	msg    = flag.String("msg", "", "alarm message")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of %s:
Either the -at or -in option must be valid.
If no -msg is given, all non-flag arguments are combined to form the message.
`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// exactly one of the arguments has to be non-empty
	if (*atTime == "") == (*inTime == "") {
		flag.Usage()
		return
	}

	if *msg == "" && len(flag.Args()) > 0 {
		*msg = strings.Join(flag.Args(), " ")
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
	start := time.Now()
	opts := win.DefaultOptions()
	opts.ClassName = "alarm_window"
	opts.Title = "Alarm"
	if *msg != "" {
		opts.Title = *msg
	}
	var handler win.MessageHandler
	window, err := win.NewWindow(opts, handler.Callback)
	if err != nil {
		panic(err)
	}
	w32.SetTimer(window, 0, 1000, 0)
	handler.OnKeyDown = func(key uintptr, _ win.KeyOptions) {
		if key == w32.VK_ESCAPE {
			win.CloseWindow(window)
		}
	}
	handler.OnTimer = func(id uintptr) {
		dur := time.Now().Sub(start)
		dur = dur - dur%time.Second
		w32.SetWindowText(window, fmt.Sprintf("%s - %v ago", opts.Title, dur))
	}
	win.RunMainLoop()
}
