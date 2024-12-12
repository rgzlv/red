package term

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

// TODO: Set raw mode similar to as documented in termios(4) for cfmakeraw().

var out = os.Stdout
var outFD = out.Fd()

type Screen struct {
	front   *Buffer
	back    *Buffer
	w, h    int
	x, y    int
	dirty   bool
	orgTerm *unix.Termios
	rawTerm *unix.Termios
	ranInit bool
	mu      sync.Mutex
}

func (s *Screen) MoveCurRel(x, y int) {
	s.x = min(s.w, max(1, s.x+x))
	s.y = min(s.h, max(1, s.y+y))
	out.Write([]byte(fmt.Sprintf("\x1B[%d;%d\x48", s.y, s.x)))
}

func (s *Screen) MoveCurAbs(x, y int) {
	s.x = min(s.w, max(1, x))
	s.y = min(s.h, max(1, y))
	out.Write([]byte(fmt.Sprintf("\x1B[%d;%d\x48", s.y, s.x)))
}

func (s *Screen) Show() {
	for _, row := range s.front.cells {
		for _, cell := range row {
			c := cell.Char
			if cell.Char == 0 {
				c = ' '
			}
			out.Write([]byte(string(c)))
		}
	}
}

func (s *Screen) Size() (w, h int) {
	return s.Width(), s.Height()
}

func (s *Screen) Width() int {
	return len(s.back.cells[0])
}

func (s *Screen) Height() int {
	return len(s.back.cells)
}

func (s *Screen) Pos() (x, y int) {
	return s.x, s.y
}

func (s *Screen) X() int {
	return s.y
}

func (s *Screen) Y() int {
	return s.x
}

func (s *Screen) Init() (err error) {
	defer func() {
		if err != nil {
			finiErr := s.Fini()
			if finiErr != nil {
				err = fmt.Errorf("%w: while cleaning up: %w", err, finiErr)
			}
		}
	}()
	if !s.ranInit {
		s.ranInit = true
		SetAlternate(true)
		s.orgTerm, err = GetTermios()
		if err != nil {
			return
		}
		rawTerm := *s.orgTerm
		s.rawTerm = &rawTerm
		SetRaw(s.rawTerm)
		err = SetTermios(s.rawTerm)
	}
	return err
}

func (s *Screen) Fini() (err error) {
	if s.ranInit {
		SetAlternate(false)
		if s.orgTerm != nil {
			err = SetTermios(s.orgTerm)
		}
		s.ranInit = false
	}
	return err
}

func NewScreen(opts ...ScreenOpt) (*Screen, error) {
	scr := &Screen{}
	winsize, err := GetWinSize()
	if err != nil {
		return nil, err
	}
	scr.w, scr.h = int(winsize.Col), int(winsize.Row)

	for _, opt := range opts {
		opt(scr)
	}

	scr.front = &Buffer{
		cells: make([][]Cell, scr.h),
	}
	scr.back = &Buffer{
		cells: make([][]Cell, scr.h),
	}
	for row := range scr.h {
		scr.front.cells[row] = make([]Cell, scr.w)
		scr.back.cells[row] = make([]Cell, scr.w)
	}

	return scr, nil
}

type ScreenOpt func(*Screen)

func ScreenWithSize(w, h int) ScreenOpt {
	return func(s *Screen) {
		s.w, s.h = w, h
	}
}

type Buffer struct {
	cells [][]Cell
	mu    sync.Mutex
}

type Cell struct {
	Format string
	Char   rune
}

func SetAlternate(enabled bool) {
	if enabled {
		out.Write([]byte("\033[?1049h"))
	} else {
		out.Write([]byte("\033[?1049l"))
	}
}

func SetRaw(term *unix.Termios) {
	term.Oflag &^= unix.OPOST
	term.Lflag &^= unix.ECHO | unix.ICANON
}

func GetTermios() (*unix.Termios, error) {
	return unix.IoctlGetTermios(int(outFD), ioctlGetTermios)
}

func SetTermios(tios *unix.Termios) error {
	return unix.IoctlSetTermios(int(outFD), ioctlSetTermios, tios)
}

func GetWinSize() (*unix.Winsize, error) {
	return unix.IoctlGetWinsize(int(outFD), unix.TIOCGWINSZ)
}
