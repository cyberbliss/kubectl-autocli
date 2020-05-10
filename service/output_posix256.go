package service

import (
	"bytes"
	"github.com/c-bata/go-prompt"

	log "github.com/sirupsen/logrus"

	"strconv"
	"syscall"
)

const flushMaxRetryCount = 3

// PosixWriter256 is a ConsoleWriter implementation for POSIX environment.
// To control terminal emulator, this outputs VT100 escape sequences.
// It is based on prompt.PosixWriter but I couldn't use prompt.VT100Writer as a promoted property
// because prompt.VT100Writer has an unexported property 'buffer' which PosixWriter256 needs access to.
// So, PosixWriter256 is a combination of both PosixWriter and VT100Writer.
// The reason for this struct is that supports the 256 colour escape sequences rather than the 16 ANSI ones

type PosixWriter256 struct {
	fd int
	buffer []byte
}

// WriteRaw to write raw byte array
func (w *PosixWriter256) WriteRaw(data []byte) {
	w.buffer = append(w.buffer, data...)
	return
}
// Flush to flush buffer
func (w *PosixWriter256) Flush() error {
	l := len(w.buffer)
	offset := 0
	retry := 0
	for {
		n, err := syscall.Write(w.fd, w.buffer[offset:])
		if err != nil {
			log.Debugf("flush error: %s", err)
			if retry < flushMaxRetryCount {
				retry++
				continue
			}
			return err
		}
		offset += n
		if offset == l {
			break
		}
	}
	w.buffer = []byte{}
	return nil
}

// Write to write safety byte array by removing control sequences.
func (w *PosixWriter256) Write(data []byte) {
	w.WriteRaw(bytes.Replace(data, []byte{0x1b}, []byte{'?'}, -1))
	return
}

// WriteRawStr to write raw string
func (w *PosixWriter256) WriteRawStr(data string) {
	w.WriteRaw([]byte(data))
	return
}

// WriteStr to write safety string by removing control sequences.
func (w *PosixWriter256) WriteStr(data string) {
	w.Write([]byte(data))
	return
}

/* Erase */

// EraseScreen erases the screen with the background colour and moves the cursor to home.
func (w *PosixWriter256) EraseScreen() {
	w.WriteRaw([]byte{0x1b, '[', '2', 'J'})
	return
}

// EraseUp erases the screen from the current line up to the top of the screen.
func (w *PosixWriter256) EraseUp() {
	w.WriteRaw([]byte{0x1b, '[', '1', 'J'})
	return
}

// EraseDown erases the screen from the current line down to the bottom of the screen.
func (w *PosixWriter256) EraseDown() {
	w.WriteRaw([]byte{0x1b, '[', 'J'})
	return
}

// EraseStartOfLine erases from the current cursor position to the start of the current line.
func (w *PosixWriter256) EraseStartOfLine() {
	w.WriteRaw([]byte{0x1b, '[', '1', 'K'})
	return
}

// EraseEndOfLine erases from the current cursor position to the end of the current line.
func (w *PosixWriter256) EraseEndOfLine() {
	w.WriteRaw([]byte{0x1b, '[', 'K'})
	return
}

// EraseLine erases the entire current line.
func (w *PosixWriter256) EraseLine() {
	w.WriteRaw([]byte{0x1b, '[', '2', 'K'})
	return
}

/* Cursor */

// ShowCursor stops blinking cursor and show.
func (w *PosixWriter256) ShowCursor() {
	w.WriteRaw([]byte{0x1b, '[', '?', '1', '2', 'l', 0x1b, '[', '?', '2', '5', 'h'})
}

// HideCursor hides cursor.
func (w *PosixWriter256) HideCursor() {
	w.WriteRaw([]byte{0x1b, '[', '?', '2', '5', 'l'})
	return
}

// CursorGoTo sets the cursor position where subsequent text will begin.
func (w *PosixWriter256) CursorGoTo(row, col int) {
	if row == 0 && col == 0 {
		// If no row/column parameters are provided (ie. <ESC>[H), the cursor will move to the home position.
		w.WriteRaw([]byte{0x1b, '[', 'H'})
		return
	}
	r := strconv.Itoa(row)
	c := strconv.Itoa(col)
	w.WriteRaw([]byte{0x1b, '['})
	w.WriteRaw([]byte(r))
	w.WriteRaw([]byte{';'})
	w.WriteRaw([]byte(c))
	w.WriteRaw([]byte{'H'})
	return
}

// CursorUp moves the cursor up by 'n' rows; the default count is 1.
func (w *PosixWriter256) CursorUp(n int) {
	if n == 0 {
		return
	} else if n < 0 {
		w.CursorDown(-n)
		return
	}
	s := strconv.Itoa(n)
	w.WriteRaw([]byte{0x1b, '['})
	w.WriteRaw([]byte(s))
	w.WriteRaw([]byte{'A'})
	return
}

// CursorDown moves the cursor down by 'n' rows; the default count is 1.
func (w *PosixWriter256) CursorDown(n int) {
	if n == 0 {
		return
	} else if n < 0 {
		w.CursorUp(-n)
		return
	}
	s := strconv.Itoa(n)
	w.WriteRaw([]byte{0x1b, '['})
	w.WriteRaw([]byte(s))
	w.WriteRaw([]byte{'B'})
	return
}

// CursorForward moves the cursor forward by 'n' columns; the default count is 1.
func (w *PosixWriter256) CursorForward(n int) {
	if n == 0 {
		return
	} else if n < 0 {
		w.CursorBackward(-n)
		return
	}
	s := strconv.Itoa(n)
	w.WriteRaw([]byte{0x1b, '['})
	w.WriteRaw([]byte(s))
	w.WriteRaw([]byte{'C'})
	return
}

// CursorBackward moves the cursor backward by 'n' columns; the default count is 1.
func (w *PosixWriter256) CursorBackward(n int) {
	if n == 0 {
		return
	} else if n < 0 {
		w.CursorForward(-n)
		return
	}
	s := strconv.Itoa(n)
	w.WriteRaw([]byte{0x1b, '['})
	w.WriteRaw([]byte(s))
	w.WriteRaw([]byte{'D'})
	return
}

// AskForCPR asks for a cursor position report (CPR).
func (w *PosixWriter256) AskForCPR() {
	// CPR: Cursor Position Request.
	w.WriteRaw([]byte{0x1b, '[', '6', 'n'})
	return
}

// SaveCursor saves current cursor position.
func (w *PosixWriter256) SaveCursor() {
	w.WriteRaw([]byte{0x1b, '[', 's'})
	return
}

// UnSaveCursor restores cursor position after a Save Cursor.
func (w *PosixWriter256) UnSaveCursor() {
	w.WriteRaw([]byte{0x1b, '[', 'u'})
	return
}

/* Scrolling */

// ScrollDown scrolls display down one line.
func (w *PosixWriter256) ScrollDown() {
	w.WriteRaw([]byte{0x1b, 'D'})
	return
}

// ScrollUp scroll display up one line.
func (w *PosixWriter256) ScrollUp() {
	w.WriteRaw([]byte{0x1b, 'M'})
	return
}

/* Title */

// SetTitle sets a title of terminal window.
func (w *PosixWriter256) SetTitle(title string) {
	titleBytes := []byte(title)
	patterns := []struct {
		from []byte
		to   []byte
	}{
		{
			from: []byte{0x13},
			to:   []byte{},
		},
		{
			from: []byte{0x07},
			to:   []byte{},
		},
	}
	for i := range patterns {
		titleBytes = bytes.Replace(titleBytes, patterns[i].from, patterns[i].to, -1)
	}

	w.WriteRaw([]byte{0x1b, ']', '2', ';'})
	w.WriteRaw(titleBytes)
	w.WriteRaw([]byte{0x07})
	return
}

// ClearTitle clears a title of terminal window.
func (w *PosixWriter256) ClearTitle() {
	w.WriteRaw([]byte{0x1b, ']', '2', ';', 0x07})
	return
}

/* Font */

// SetColor sets text and background colors. and specify whether text is bold.
func (w *PosixWriter256) SetColor(fg, bg prompt.Color, bold bool) {
	if bold {
		w.SetDisplayAttributes(fg, bg, prompt.DisplayBold)
	} else {
		// If using `DisplayDefualt`, it will be broken in some environment.
		// Details are https://github.com/c-bata/go-prompt/pull/85
		w.SetDisplayAttributes(fg, bg, prompt.DisplayReset)
	}
	return
}

// SetDisplayAttributes to set VT100 display attributes.
func (w *PosixWriter256) SetDisplayAttributes(fg, bg prompt.Color, attrs ...prompt.DisplayAttribute) {
	w.WriteRaw([]byte{0x1b, '['}) // control sequence introducer
	defer w.WriteRaw([]byte{'m'}) // final character

	var separator byte = ';'
	for i := range attrs {
		p, ok := displayAttributeParameters[attrs[i]]
		if !ok {
			continue
		}
		w.WriteRaw(p)
		w.WriteRaw([]byte{separator})
	}

	var f, b []byte

	// if the fg or bg value is 0 this means use the default colour for the terminal
	if fg == 0 {
		f = []byte{'3', '9'}
	} else {
		f = []byte{'3', '8', separator, '5', separator}
		f = append(f, Color2Byte(fg)...)
	}

	if bg == 0 {
		b = []byte{'4', '9'}
	} else {
		b = []byte{'4','8',separator,'5',separator}
		b = append(b, Color2Byte(bg)...)
	}

	w.WriteRaw(f)
	w.WriteRaw([]byte{separator})
	w.WriteRaw(b)

	return
}

var displayAttributeParameters = map[prompt.DisplayAttribute][]byte{
	prompt.DisplayReset:        {'0'},
	prompt.DisplayBold:         {'1'},
	prompt.DisplayLowIntensity: {'2'},
	prompt.DisplayItalic:       {'3'},
	prompt.DisplayUnderline:    {'4'},
	prompt.DisplayBlink:        {'5'},
	prompt.DisplayRapidBlink:   {'6'},
	prompt.DisplayReverse:      {'7'},
	prompt.DisplayInvisible:    {'8'},
	prompt.DisplayCrossedOut:   {'9'},
	prompt.DisplayDefaultFont:  {'1', '0'},
}

//var foregroundANSIColors = map[prompt.Color][]byte{
//	prompt.DefaultColor: {'3', '9'},
//
//	// Low intensity.
//	prompt.Black:     {'3', '0'},
//	prompt.DarkRed:   {'3', '1'},
//	prompt.DarkGreen: {'3', '2'},
//	prompt.Brown:     {'3', '3'},
//	prompt.DarkBlue:  {'3', '4'},
//	prompt.Purple:    {'3', '5'},
//	prompt.Cyan:      {'3', '6'},
//	prompt.LightGray: {'3', '7'},
//
//	// High intensity.
//	prompt.DarkGray:  {'9', '0'},
//	prompt.Red:       {'9', '1'},
//	prompt.Green:     {'9', '2'},
//	prompt.Yellow:    {'9', '3'},
//	prompt.Blue:      {'9', '4'},
//	prompt.Fuchsia:   {'9', '5'},
//	prompt.Turquoise: {'9', '6'},
//	prompt.White:     {'9', '7'},
//}
//
//var backgroundANSIColors = map[prompt.Color][]byte{
//	prompt.DefaultColor: {'4', '9'},
//
//	// Low intensity.
//	prompt.Black:     {'4', '0'},
//	prompt.DarkRed:   {'4', '1'},
//	prompt.DarkGreen: {'4', '2'},
//	prompt.Brown:     {'4', '3'},
//	prompt.DarkBlue:  {'4', '4'},
//	prompt.Purple:    {'4', '5'},
//	prompt.Cyan:      {'4', '6'},
//	prompt.LightGray: {'4', '7'},
//
//	// High intensity
//	prompt.DarkGray:  {'1', '0', '0'},
//	prompt.Red:       {'1', '0', '1'},
//	prompt.Green:     {'1', '0', '2'},
//	prompt.Yellow:    {'1', '0', '3'},
//	prompt.Blue:      {'1', '0', '4'},
//	prompt.Fuchsia:   {'1', '0', '5'},
//	prompt.Turquoise: {'1', '0', '6'},
//	prompt.White:     {'1', '0', '7'},
//}


var _ prompt.ConsoleWriter = &PosixWriter256{}

// NewStdoutWriter returns ConsoleWriter object to write to stdout.
// This generates VT100 escape sequences because almost terminal emulators
// in POSIX OS built on top of a VT100 specification.
func NewStdoutWriter() prompt.ConsoleWriter {
	return &PosixWriter256{
		fd: syscall.Stdout,
	}
}

// NewStderrWriter returns ConsoleWriter object to write to stderr.
// This generates VT100 escape sequences because almost terminal emulators
// in POSIX OS built on top of a VT100 specification.
func NewStderrWriter() prompt.ConsoleWriter {
	return &PosixWriter256{
		fd: syscall.Stderr,
	}
}

func Color2Byte(src prompt.Color) []byte {
	return []byte(strconv.Itoa(int(src)))
}

