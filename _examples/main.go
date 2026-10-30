// Command _examples draws the color table of the terminal, and waits for
// Ctrl-C.
package main

import (
	"bytes"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/xo/terminfo"
)

func main() {
	// load terminfo
	ti, err := terminfo.LoadFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	// cleanup
	defer func() {
		err := recover()
		termreset(ti)
		if err != nil {
			log.Fatal(err)
		}
	}()
	terminit(ti)
	termtitle(ti, "simple example!")
	termputs(ti, 3, 3, "Ctrl-C to exit")
	for i := range min(termcolors(ti), 256) {
		termputs(ti, 5+i/16, 5+i%16, ti.Colorf(i, 0, "█"))
	}
	// wait for signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

// terminit initializes the special CA mode on the terminal, and makes the
// cursor invisible.
func terminit(ti *terminfo.Terminfo) {
	buf := new(bytes.Buffer)
	// set the cursor invisible
	ti.Fprintf(buf, terminfo.CursorInvisible)
	// enter special mode
	ti.Fprintf(buf, terminfo.EnterCaMode)
	// clear the screen
	ti.Fprintf(buf, terminfo.ClearScreen)
	_, _ = os.Stdout.Write(buf.Bytes())
}

// termreset is the inverse of terminit.
func termreset(ti *terminfo.Terminfo) {
	buf := new(bytes.Buffer)
	ti.Fprintf(buf, terminfo.ExitCaMode)
	ti.Fprintf(buf, terminfo.CursorNormal)
	_, _ = os.Stdout.Write(buf.Bytes())
}

// termputs writes s at row, col.
func termputs(ti *terminfo.Terminfo, row, col int, s string) {
	buf := new(bytes.Buffer)
	ti.Fprintf(buf, terminfo.CursorAddress, row, col)
	buf.WriteString(s)
	_, _ = os.Stdout.Write(buf.Bytes())
}

// sl is the status line terminfo, and slOnce loads it.
var (
	sl     *terminfo.Terminfo
	slOnce sync.Once
)

// termtitle sets the window title.
func termtitle(ti *terminfo.Terminfo, s string) {
	slOnce.Do(func() {
		if ti.Has(terminfo.HasStatusLine) {
			return
		}
		// load the sl xterm if terminal is an xterm or has COLORTERM
		if strings.Contains(strings.ToLower(os.Getenv("TERM")), "xterm") || os.Getenv("COLORTERM") == "truecolor" {
			sl, _ = terminfo.Load("xterm+sl")
		}
	})
	if sl != nil {
		ti = sl
	}
	if !ti.Has(terminfo.HasStatusLine) {
		return
	}
	buf := new(bytes.Buffer)
	ti.Fprintf(buf, terminfo.ToStatusLine)
	buf.WriteString(s)
	ti.Fprintf(buf, terminfo.FromStatusLine)
	_, _ = os.Stdout.Write(buf.Bytes())
}

// termcolors returns the maximum colors available for the terminal.
func termcolors(ti *terminfo.Terminfo) int {
	if colors := ti.Num(terminfo.MaxColors); colors > 0 {
		return colors
	}
	return int(terminfo.ColorLevelBasic)
}
