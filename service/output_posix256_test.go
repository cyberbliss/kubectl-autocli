package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVT100WriterWrite(t *testing.T) {
	scenarioTable := []struct {
		input    []byte
		expected []byte
	}{
		{
			input:    []byte{0x1b},
			expected: []byte{'?'},
		},
		{
			input:    []byte{'a'},
			expected: []byte{'a'},
		},
	}

	for _, s := range scenarioTable {
		pw := &PosixWriter256{}
		pw.Write(s.input)
		assert.Equal(t, s.expected, pw.buffer)
	}
}

func TestVT100WriterWriteStr(t *testing.T) {
	scenarioTable := []struct {
		input    string
		expected []byte
	}{
		{
			input:    "\x1b",
			expected: []byte{'?'},
		},
		{
			input:    "a",
			expected: []byte{'a'},
		},
	}

	for _, s := range scenarioTable {
		pw := &PosixWriter256{}
		pw.WriteStr(s.input)

		assert.Equal(t, s.expected, pw.buffer)
	}
}

func TestVT100WriterWriteRawStr(t *testing.T) {
	scenarioTable := []struct {
		input    string
		expected []byte
	}{
		{
			input:    "\x1b",
			expected: []byte{0x1b},
		},
		{
			input:    "a",
			expected: []byte{'a'},
		},
	}

	for _, s := range scenarioTable {
		pw := &PosixWriter256{}
		pw.WriteRawStr(s.input)

		assert.Equal(t, s.expected, pw.buffer)
	}
}
