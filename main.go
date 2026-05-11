package main

import (
	"io"
	"os"
	"os/exec"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[?=0-9;]*[a-zA-Z]|\x1b\](?:[^\x07\x1b]*)(?:\x07|\x1b\\)|[\x00-\x08\x0b\x0c\x0e-\x1f]`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func main() {
	a := app.New()
	w := a.NewWindow("goterm")

	ui := widget.NewTextGrid()

	c := exec.Command("/bin/bash")
	p, err := pty.Start(c)
	if err != nil {
		fyne.LogError("Failed to open pty", err)
		os.Exit(1)
	}

	defer c.Process.Kill()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := p.Read(buf)
			if err != nil {
				if err != io.EOF {
					fyne.LogError("PTY read error", err)
				}
				return
			}
			clean := stripAnsi(string(buf[:n]))
			fyne.Do(func() {
				ui.SetText(ui.Text() + clean)
			})
		}
	}()

	w.Canvas().SetOnTypedRune(func(r rune) {
		p.Write([]byte(string(r)))
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		switch e.Name {
		case fyne.KeyEnter, fyne.KeyReturn:
			p.Write([]byte("\r"))
		case fyne.KeyBackspace:
			p.Write([]byte{127})
		case fyne.KeyUp:
			p.Write([]byte("\x1b[A"))
		case fyne.KeyDown:
			p.Write([]byte("\x1b[B"))
		case fyne.KeyLeft:
			p.Write([]byte("\x1b[D"))
		case fyne.KeyRight:
			p.Write([]byte("\x1b[C"))
		case fyne.KeyTab:
			p.Write([]byte("\t"))
		case fyne.KeyEscape:
			p.Write([]byte("\x1b"))
		}
	})
	w.SetContent(container.NewMax(ui))
	w.Resize(fyne.NewSize(800, 500))
	w.ShowAndRun()
}
