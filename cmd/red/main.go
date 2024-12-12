package main

import (
	"fmt"
	"log"
	"os"
	"red/event"
	"red/term"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("red: ")
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	var scr *term.Screen
	scr, err = term.NewScreen()
	if err != nil {
		return
	}
	defer func() {
		if finiErr := scr.Fini(); finiErr != nil {
			err = fmt.Errorf("fini: %w: %w", err, finiErr)
		}
	}()
	if err = scr.Init(); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	scr.Show()
	scr.MoveCurAbs(1, 1)

	var keyListener event.KeyListener
	keyCh, err := keyListener.Listen(os.Stdin)
	if err != nil {
		return err
	}
loop:
	for {
		select {
		case key := <-keyCh:
			if key.Err != nil {
				return err
			}
			if key.Char == 'q' {
				break loop
			}
			handleKey(scr, key)
		}
	}

	return err
}

func handleKey(scr *term.Screen, key event.KeyEvent) {
	switch key.Char {
	case 'h':
		scr.MoveCurRel(-1, 0)
	case 'j':
		scr.MoveCurRel(0, 1)
	case 'k':
		scr.MoveCurRel(0, -1)
	case 'l':
		scr.MoveCurRel(1, 0)
	}
}
