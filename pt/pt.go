package pt

import (
	"errors"
	"fmt"
	"strings"
)

type Editor struct {
	Org   string
	Add   string
	Table *Table
	Ops   Ops
}

func (e *Editor) Insert(pos int, s string) error {
	op, err := e.Table.Insert(pos, len(s), len(e.Add))
	if err != nil {
		return err
	}
	e.Add += s
	e.Ops = e.Ops.Push(op)
	return nil
}

func (e *Editor) MustInsert(pos int, s string) {
	err := e.Insert(pos, s)
	if err != nil {
		panic(err)
	}
}

func (e *Editor) Seq() string {
	return e.Table.Seq(e.Org, e.Add)
}

func (e *Editor) Undo() {
	var op Op
	e.Ops, op = e.Ops.Pop()
	if op != nil {
		op.Undo()
	}
}

func (e *Editor) Format(f fmt.State, verb rune) {
	w, _ := f.Width()
	ws := strings.Repeat(" ", w)

	tableFmt := "%"
	fields := map[string]any{
		"Org": e.Org,
		"Add": e.Add,
	}
	keys := []string{"Org", "Add"}
	if f.Flag('#') {
		fields["Table"] = e.Table
		keys = append(keys, "Table")
		tableFmt += "#"
	}
	if f.Flag('+') {
		fields["Ops"] = e.Ops
		keys = append(keys, "Ops")
		tableFmt += "+"
	}
	tableFmt += "*v"

	for _, key := range keys {
		if val, ok := fields[key]; ok {
			if tbl, ok := val.(*Table); ok {
				f.Write([]byte(fmt.Sprintf("%sTable:\n", ws)))
				f.Write([]byte(fmt.Sprintf(tableFmt, w+2, tbl)))
				continue
			}
			if ops, ok := val.(Ops); ok {
				f.Write([]byte(fmt.Sprintf("%sOps:\n", ws)))
				for _, op := range ops {
					var opStr string
					switch op.(type) {
					case *OpInsert:
						opStr = "Insert"
					default:
						opStr = "Unrecognized"
					}
					f.Write([]byte(fmt.Sprintf("%*s:\n", len(opStr)+w+2, opStr)))
					f.Write([]byte(fmt.Sprintf("%*v", w+4, op)))
				}
				continue
			}
			f.Write([]byte(fmt.Sprintf("%s%s: %v\n", ws, key, val)))
		}
	}
}

func NewEditor(org string) *Editor {
	e := &Editor{
		Org:   org,
		Table: NewTable(len(org)),
	}
	return e
}

type Table struct {
	Head *Piece
	Len  int
}

func (t *Table) Insert(pos, ln, addBufPos int) (Op, error) {
	var found bool
	op := &OpInsert{}
	t.ForEach(func(p *Piece, off int) bool {
		if !(pos >= off && pos <= off+p.Len) {
			return true
		}
		found = true

		op.Pre = *p
		ins := &Piece{
			Buf: BufAdd,
			Pos: addBufPos,
			Len: ln,
		}
		switch {
		case pos == off:
			op.Dir = InsertDirLeft
			ins.Prev, ins.Next = p.Prev, p
			p.Prev.Next = ins
			p.Prev = ins
			t.Len += 1
		case pos == off+p.Len:
			op.Dir = InsertDirRight
			ins.Prev, ins.Next = p, p.Next
			p.Next.Prev = ins
			p.Next = ins
			t.Len += 1
		default:
			op.Dir = InsertDirMid
			oldlen := p.Len
			p.Len = pos - off
			rhs := &Piece{
				Buf: p.Buf,
				Pos: p.Pos + p.Len,
				Len: oldlen - (pos - off),
			}
			rhs.Prev, rhs.Next = ins, p.Next
			ins.Prev, ins.Next = p, rhs
			p.Next.Prev = rhs
			p.Next = ins
			t.Len += 2
		}

		return false
	})
	if !found {
		return nil, errors.New("couldn't find piece in which to insert")
	}
	return op, nil
}

func (t *Table) Seq(org, add string) string {
	var s string
	t.ForEach(func(p *Piece) {
		var buf string
		if p.Buf == BufOrg {
			buf = org
		} else {
			buf = add
		}
		s += buf[p.Pos : p.Pos+p.Len]
	})
	return s
}

func (t *Table) ForEach(fn any) int {
	var idx int
	var off int
loop:
	for p := t.Head.Next; p != t.Head; p = p.Next {
		switch fn := fn.(type) {
		case func():
			fn()
		case func() bool:
			if !fn() {
				break loop
			}
		case func(*Piece):
			fn(p)
		case func(*Piece) bool:
			if !fn(p) {
				break loop
			}
		case func(*Piece, int):
			fn(p, off)
		case func(*Piece, int) bool:
			if !fn(p, off) {
				break loop
			}
		case func(*Piece, int, int):
			fn(p, off, idx)
		case func(*Piece, int, int) bool:
			if !fn(p, off, idx) {
				break loop
			}
		default:
			panic("unrecognized function")
		}

		idx++
		off += p.Len
	}
	return idx
}

func (t *Table) Format(f fmt.State, verb rune) {
	w, _ := f.Width()
	ws := strings.Repeat(" ", w)
	pieceFmt := "%"
	if f.Flag('#') {
		pieceFmt += "#"
	}
	if f.Flag('+') {
		pieceFmt += "+"
	}
	pieceFmt += "*v"

	f.Write([]byte(fmt.Sprintf("%sLen: %v\n", ws, t.Len)))
	f.Write([]byte(fmt.Sprintf("%sHead: %p\n", ws, t.Head)))
	t.ForEach(func(p *Piece, _ int, idx int) {
		f.Write([]byte(fmt.Sprintf("%*sPiece[%d]:\n", w+2, "", idx)))
		f.Write([]byte(fmt.Sprintf(pieceFmt, w+2, p)))
	})
}

func NewTable(orgLen int) *Table {
	t := &Table{
		Len: 1,
	}
	orgPiece := &Piece{
		Buf: BufOrg,
		Pos: 0,
		Len: orgLen,
	}
	t.Head = &Piece{
		Prev: orgPiece,
		Next: orgPiece,
	}
	orgPiece.Prev, orgPiece.Next = t.Head, t.Head
	return t
}

type Piece struct {
	Buf Buf
	Pos int
	Len int

	Prev, Next *Piece
}

func (p *Piece) Format(f fmt.State, verb rune) {
	w, _ := f.Width()
	ws := strings.Repeat(" ", w)

	fields := map[string]any{
		"Buf": p.Buf,
		"Pos": p.Pos,
		"Len": p.Len,
	}
	keys := []string{"Buf", "Pos", "Len"}
	if f.Flag('#') {
		fields["This"] = p
		keys = append(keys, "This")
	}
	if f.Flag('+') {
		fields["Prev"] = p.Prev
		fields["Next"] = p.Next
		keys = append(keys, "Prev", "Next")
	}

	for _, key := range keys {
		if val, ok := fields[key]; ok {
			if key == "This" || key == "Prev" || key == "Next" {
				f.Write([]byte(fmt.Sprintf("%s%s: %p\n", ws, key, val)))
			} else {
				f.Write([]byte(fmt.Sprintf("%s%s: %v\n", ws, key, val)))
			}
		}
	}
}

type Buf bool

const (
	BufOrg Buf = true
	BufAdd Buf = false
)

func (b Buf) String() string {
	if b {
		return "BufOrg"
	}
	return "BufAdd"
}

type Ops []Op

func (ops Ops) Undo() {
}

func (ops Ops) Push(op Op) Ops {
	return append(ops, op)
}

func (ops Ops) Pop() (Ops, Op) {
	if len(ops) > 0 {
		op := ops[len(ops)-1]
		return ops[:len(ops)-1], op
	}
	return nil, nil
}

type Op interface {
	Undo()
}

type OpInsert struct {
	Pre Piece
	Dir InsertDir
}

func (op *OpInsert) Undo() {
	switch op.Dir {
	case InsertDirMid:
		cpyPre := op.Pre
		op.Pre.Prev.Next = &cpyPre
		op.Pre.Next.Prev = &cpyPre
	default:
		panic("TODO: impl other insert dir undos")
	}
}

func (op *OpInsert) Format(f fmt.State, verb rune) {
	w, _ := f.Width()
	ws := strings.Repeat(" ", w)

	f.Write([]byte(fmt.Sprintf("%sPre:\n", ws)))
	f.Write([]byte(fmt.Sprintf("%*v", w+2, &op.Pre)))
	f.Write([]byte(fmt.Sprintf("%sDir: %v\n", ws, op.Dir)))
}

//go:generate stringer -type InsertDir
type InsertDir int

const (
	InsertDirLeft InsertDir = iota
	InsertDirMid
	InsertDirRight
)
