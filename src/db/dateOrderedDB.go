package db

import (
	"bytes"
)

type DateComparator struct{}

func (d *DateComparator) Compare(a, b []byte) int {
	return bytes.Compare(b, a)
}

func (d *DateComparator) Name() string {
	return "dateReverseComparer"
}

func (d *DateComparator) Equal(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func (d *DateComparator) AbbreviatedKey(key []byte) uint64 {
	// This is a simplistic implementation. Adjust as needed.
	if len(key) == 0 {
		return 0
	}
	return uint64(key[0])
}

func (d *DateComparator) FormatKey(k []byte) string {
	return string(k)
}

func (d *DateComparator) FormatValue(v []byte) string {
	return string(v)
}

func (d *DateComparator) Separator(dst, a, b []byte) []byte {
	return a
}

func (d *DateComparator) Split(dst, a, b []byte) []byte {
	return b
}

func (d *DateComparator) Successor(dst, a []byte) []byte {
	return a // simplistic stub, adjust as needed
}

func (d *DateComparator) ImmediateSuccessor(dst, k []byte) []byte {
	return k // simplistic stub, adjust as needed
}
