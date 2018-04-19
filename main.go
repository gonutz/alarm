package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
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
		fmt.Scanln()
		return
	}

	if *msg == "" && len(flag.Args()) > 0 {
		*msg = strings.Join(flag.Args(), " ")
	}

	var waitTime time.Duration
	if *atTime != "" {
		at, err := parseTime(*atTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing time at which to ring the alarm:", err)
			flag.Usage()
			fmt.Scanln()
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
			fmt.Scanln()
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
	handler.OnTimer = func(id uintptr) {
		dur := time.Now().Sub(start)
		dur = dur - dur%time.Second
		w32.SetWindowText(window, fmt.Sprintf("%s - %v ago", opts.Title, dur))
	}
	win.RunMainLoop()
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.Kitchen, s)
	if err == nil {
		return t, nil
	}
	var hStr, minStr string
	if strings.Contains(s, ":") {
		parts := strings.SplitN(s, ":", 2)
		hStr = parts[0]
		minStr = parts[1]
	} else {
		hStr = s
		minStr = "0"
	}
	h, err := strconv.Atoi(hStr)
	if err != nil {
		return time.Time{}, err
	}
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return time.Time{}, err
	}
	if !(0 <= h && h <= 23) {
		return time.Time{}, errors.New("hours must be in the range [0..23]")
	}
	if !(0 <= min && min <= 59) {
		return time.Time{}, errors.New("minutes must be in the range [0..59]")
	}
	ampm := "AM"
	if h == 12 {
		ampm = "PM"
	}
	if h >= 13 {
		h -= 12
		ampm = "PM"
	}
	return time.Parse(time.Kitchen, fmt.Sprintf(
		"%d%d:%d%d%s",
		h/10, h%10, min/10, min%10, ampm,
	))
}
