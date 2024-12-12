package pt

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTableNew(t *testing.T) {
	e := NewEditor("ABCDEF")
	expectEqPiece(t, e.Table.Head.Next, &Piece{
		Buf:  BufOrg,
		Pos:  0,
		Len:  6,
		Prev: e.Table.Head,
		Next: e.Table.Head,
	})
	expectEq(t, e.Seq(), "ABCDEF")
}

func TestTableInsert(t *testing.T) {
	type insertCase struct {
		pos int
		s   string
	}
	testCases := map[string]struct {
		org    string
		seq    string
		insert []insertCase
		expect []*Piece
	}{
		"once": {
			seq:    "ABCDXEF",
			insert: []insertCase{{4, "X"}},
			expect: []*Piece{
				{
					Buf: BufOrg,
					Pos: 0,
					Len: 4,
				}, // "ABCD"
				{
					Buf: BufAdd,
					Pos: 0,
					Len: 1,
				}, // "X"
				{
					Buf: BufOrg,
					Pos: 4,
					Len: 2,
				}, // "EF"
			},
		},
		"multiple": {
			seq:    "ABXCDYEF",
			insert: []insertCase{{2, "X"}, {5, "Y"}},
			expect: []*Piece{
				{
					Buf: BufOrg,
					Pos: 0,
					Len: 2,
				}, // "AB"
				{
					Buf: BufAdd,
					Pos: 0,
					Len: 1,
				}, // "X"
				{
					Buf: BufOrg,
					Pos: 2,
					Len: 2,
				}, // "CD"
				{
					Buf: BufAdd,
					Pos: 1,
					Len: 1,
				}, // "Y"
				{
					Buf: BufOrg,
					Pos: 4,
					Len: 2,
				}, // "EF"
			},
		},
		"overlapping": {
			seq:    "ABXYXCDEF",
			insert: []insertCase{{2, "XX"}, {3, "Y"}},
			expect: []*Piece{
				{
					Buf: BufOrg,
					Pos: 0,
					Len: 2,
				}, // "AB"
				{
					Buf: BufAdd,
					Pos: 0,
					Len: 1,
				}, // "X"
				{
					Buf: BufAdd,
					Pos: 2, // Not 1, "XX" appended before "Y".
					Len: 1,
				}, // "Y"
				{
					Buf: BufAdd,
					Pos: 1,
					Len: 1,
				}, // "X"
				{
					Buf: BufOrg,
					Pos: 2,
					Len: 4,
				}, // "CDEF"
			},
		},
		"beginning": {
			seq:    "XABCDEF",
			insert: []insertCase{{0, "X"}},
			expect: []*Piece{
				{
					Buf: BufAdd,
					Pos: 0,
					Len: 1,
				},
				{
					Buf: BufOrg,
					Pos: 0,
					Len: 6,
				},
			},
		},
		"end": {
			seq:    "ABCDEFX",
			insert: []insertCase{{6, "X"}},
			expect: []*Piece{
				{
					Buf: BufOrg,
					Pos: 0,
					Len: 6,
				},
				{
					Buf: BufAdd,
					Pos: 0,
					Len: 1,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if testCase.org == "" {
				testCase.org = "ABCDEF"
			}
			e := NewEditor(testCase.org)

			for _, insert := range testCase.insert {
				if err := e.Insert(insert.pos, insert.s); err != nil {
					t.Fatalf("couldn't insert: %v", err)
				}
			}

			expectEqPieces(t, e.Table, testCase.expect...)

			if testCase.seq != "" {
				expectEq(t, e.Seq(), testCase.seq)
			}

			var last *Piece
			e.Table.ForEach(func(p *Piece) {
				last = p
			})
			if e.Table.Head.Prev != last {
				t.Fatal("table Head.Prev not pointing to last piece")
			}
		})
	}
}

func expectEq(t *testing.T, a, b any) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expected equal values, got %v != %b", a, b)
	}
}

func expectEqPieces(t *testing.T, tbl *Table, target ...*Piece) {
	t.Helper()
	var index int
	tbl.ForEach(func(p *Piece) {
		if len(target) == 0 {
			t.Fatalf("end of target reached at index %d", index)
		}
		if err := eqPiece(p, target[0]); err != nil {
			t.Fatalf("piece[%d] != target[%[1]d]: %v", index, err)
		}
		index++
		target = target[1:]
	})
}

func expectEqPiece(t *testing.T, piece, target *Piece) {
	t.Helper()
	if err := eqPiece(piece, target); err != nil {
		t.Fatalf("expected equal pieces: %v", err)
	}
}

func eqPiece(piece, target *Piece) error {
	var err error
	switch {
	case piece == nil:
		err = fmt.Errorf("piece is nil")
	case target == nil:
		err = fmt.Errorf("target is nil")
	case piece.Buf != target.Buf:
		err = fmt.Errorf("piece.Buf (%v) != target.Buf (%v)", piece.Buf, target.Buf)
	case piece.Pos != target.Pos:
		err = fmt.Errorf("piece.Pos (%v) != target.Pos (%v)", piece.Pos, target.Pos)
	case piece.Len != target.Len:
		err = fmt.Errorf("piece.Len (%v) != target.Len (%v)", piece.Len, target.Len)
	}
	return err
}
