package event

import (
	"errors"
	"io"
	"unicode/utf8"
)

type KeyListener struct {
	ch  chan<- KeyEvent
	buf []byte
	src io.Reader
}

func (l *KeyListener) Listen(src io.Reader) (<-chan KeyEvent, error) {
	ch := make(chan KeyEvent)
	if l.ch == nil {
		l.ch = ch
	}
	if l.buf == nil {
		l.buf = make([]byte, 4)
	}
	if src != nil {
		l.src = src
	}
	if l.src == nil {
		return nil, errors.New("nil input source")
	}
	go l.listen()
	return ch, nil
}

func (l *KeyListener) listen() {
	for {
		ev := KeyEvent{}
		_, ev.Err = l.src.Read(l.buf)
		if ev.Err == nil {
			ev.Char, _ = utf8.DecodeRune(l.buf)
		}
		clear(l.buf)
		l.ch <- ev
	}
}

type KeyEvent struct {
	Char rune
	Err  error
}
