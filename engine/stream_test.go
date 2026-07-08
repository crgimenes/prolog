package engine

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestNewInputTextStream(t *testing.T) {
	equal(t, &Stream{
		source:     os.Stdin,
		mode:       ioModeRead,
		eofAction:  eofActionReset,
		streamType: streamTypeText,
	}, NewInputTextStream(os.Stdin))
}

func TestNewInputBinaryStream(t *testing.T) {
	equal(t, &Stream{
		source:     os.Stdin,
		mode:       ioModeRead,
		eofAction:  eofActionReset,
		streamType: streamTypeBinary,
	}, NewInputBinaryStream(os.Stdin))
}

func TestNewOutputTextStream(t *testing.T) {
	equal(t, &Stream{
		sink:       os.Stdout,
		mode:       ioModeAppend,
		eofAction:  eofActionReset,
		streamType: streamTypeText,
	}, NewOutputTextStream(os.Stdout))
}

func TestNewOutputBinaryStream(t *testing.T) {
	equal(t, &Stream{
		sink:       os.Stdout,
		mode:       ioModeAppend,
		eofAction:  eofActionReset,
		streamType: streamTypeBinary,
	}, NewOutputBinaryStream(os.Stdout))
}

func TestStream_WriteTerm(t *testing.T) {
	tests := []struct {
		title  string
		s      *Stream
		opts   WriteOptions
		output string
	}{
		{title: "stream", s: &Stream{}, output: `<stream>\(0x[[:xdigit:]]+\)`},
	}

	var buf bytes.Buffer
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			buf.Reset()
			noError(t, tt.s.WriteTerm(&buf, &tt.opts, nil))
			matchRegexp(t, tt.output, buf.String())
		})
	}
}

func TestStream_Compare(t *testing.T) {
	x := NewVariable()
	var ss [3]Stream

	tests := []struct {
		title string
		s     *Stream
		t     Term
		o     int
	}{
		{title: `s > X`, s: &ss[1], t: x, o: 1},
		{title: `s > 1.0`, s: &ss[1], t: Float(1), o: 1},
		{title: `s > 1`, s: &ss[1], t: Integer(2), o: 1},
		{title: `s > a`, s: &ss[1], t: NewAtom("a"), o: 1},
		{title: `s > s`, s: &ss[1], t: &ss[0], o: 1},
		{title: `s = s`, s: &ss[1], t: &ss[1], o: 0},
		{title: `s < s`, s: &ss[1], t: &ss[2], o: -1},
		{title: `s < f(a)`, s: &ss[1], t: NewAtom("f").Apply(NewAtom("a")), o: -1},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			equal(t, tt.o, tt.s.Compare(tt.t, nil))
		})
	}
}

type mockNamer struct {
	name string
}

func (m *mockNamer) Name() string {
	return m.name
}

func TestStream_Name(t *testing.T) {
	t.Run("namer", func(t *testing.T) {
		var m struct {
			mockReader
			mockNamer
		}
		m.name = "name"

		s := &Stream{source: &m}
		equal(t, "name", s.Name())
	})

	t.Run("not namer", func(t *testing.T) {
		var m mockWriter

		s := &Stream{sink: &m}
		equal(t, "", s.Name())
	})
}

// mockFile is a configurable os.File-like source. Stat, Read, and Seek each
// return what the test sets; the zero value is an empty, error-free file.
type mockFile struct {
	readData string
	statErr  error
	seekErr  error
}

func (m *mockFile) Stat() (fs.FileInfo, error) {
	return &mockFileInfo{}, m.statErr
}

func (m *mockFile) Read(p []byte) (int, error) {
	return copy(p, m.readData), nil
}

func (m *mockFile) Close() error {
	return nil
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	return 0, m.seekErr
}

// mockFileInfo satisfies fs.FileInfo. The stream only ever inspects it after a
// successful Stat, which the tests here never arrange, so the methods return
// zero values.
type mockFileInfo struct{}

func (mockFileInfo) Name() string      { return "" }
func (mockFileInfo) Size() int64       { return 0 }
func (mockFileInfo) Mode() fs.FileMode { return 0 }
func (mockFileInfo) ModTime() time.Time {
	return time.Time{}
}
func (mockFileInfo) IsDir() bool { return false }
func (mockFileInfo) Sys() any    { return nil }

type mockCloser struct {
	err error
}

func (m *mockCloser) Close() error {
	return m.err
}

func TestStream_Close(t *testing.T) {
	var okCloser struct {
		mockReader
		mockCloser
	}

	var ngCloser struct {
		mockWriter
		mockCloser
	}
	ngCloser.mockCloser.err = errors.New("ng")

	var vm VM

	foo := NewAtom("foo")
	s := &Stream{vm: &vm, source: &okCloser, alias: foo}
	vm.streams.add(s)

	bar := NewAtom("bar")
	vm.streams.add(&Stream{vm: &vm, source: &okCloser, alias: bar})

	tests := []struct {
		title string
		s     *Stream
		err   error
	}{
		{title: "ok closer", s: &Stream{source: &okCloser}},
		{title: "not closer", s: &Stream{source: &okCloser.mockReader}},
		{title: "alias", s: s},

		{title: "ng closer", s: &Stream{sink: &ngCloser}, err: errors.New("ng")},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			equal(t, tt.err, tt.s.Close())
		})
	}
}

// mockReader is a configurable io.Reader used to inject read failures. Where a
// test needs a source that is not a Closer or Seeker, the zero value serves as
// a bare reader whose Read is never reached.
type mockReader struct {
	n   int
	err error
}

func (m *mockReader) Read(p []byte) (int, error) {
	return m.n, m.err
}

func TestStream_ReadByte(t *testing.T) {
	tests := []struct {
		title string
		s     *Stream
		b     byte
		err   error
		pos   int64
		eos   endOfStream
	}{
		{
			title: "input binary: 3 bytes left",
			s:     &Stream{source: bytes.NewReader([]byte{1, 2, 3}), streamType: streamTypeBinary},
			b:     1,
			pos:   1,
			eos:   endOfStreamNot,
		},
		{
			title: "input binary: 2 bytes left",
			s:     &Stream{source: bytes.NewReader([]byte{2, 3}), streamType: streamTypeBinary, position: 1},
			b:     2,
			pos:   2,
			eos:   endOfStreamNot,
		},
		{
			title: "input binary: 1 byte left",
			s:     &Stream{source: bytes.NewReader([]byte{3}), streamType: streamTypeBinary, position: 2},
			b:     3,
			pos:   3,
			eos:   endOfStreamNot,
		},
		{
			title: "input binary: empty",
			s:     &Stream{source: bytes.NewReader([]byte{}), streamType: streamTypeBinary, position: 3},
			err:   io.EOF,
			pos:   3,
			eos:   endOfStreamPast,
		},
		{
			title: "end of stream past: error",
			s:     &Stream{source: bytes.NewReader([]byte{1, 2, 3}), streamType: streamTypeBinary, endOfStream: endOfStreamPast, eofAction: eofActionError},
			err:   errPastEndOfStream,
			pos:   0,
			eos:   endOfStreamPast,
		},
		{
			title: "end of stream past: reset",
			s:     &Stream{source: bytes.NewReader([]byte{1, 2, 3}), streamType: streamTypeBinary, endOfStream: endOfStreamPast, eofAction: eofActionReset, reposition: true},
			b:     1,
			pos:   1,
			eos:   endOfStreamNot,
		},
		{
			title: "input text",
			s:     &Stream{source: bytes.NewReader([]byte{1, 2, 3}), streamType: streamTypeText},
			err:   errWrongStreamType,
		},
		{
			title: "output",
			s:     &Stream{source: bytes.NewReader([]byte{1, 2, 3}), mode: ioModeAppend},
			err:   errWrongIOMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			b, err := tt.s.ReadByte()
			equal(t, tt.b, b)
			equal(t, tt.err, err)

			equal(t, tt.pos, tt.s.position)
			equal(t, tt.eos, tt.s.endOfStream)
		})
	}
}

func TestStream_ReadRune(t *testing.T) {
	var m mockFile
	m.statErr = errors.New("failed")
	m.readData = "a"

	tests := []struct {
		title string
		s     *Stream
		r     rune
		size  int
		err   error
		pos   int64
		eos   endOfStream
	}{
		{
			title: "input text: 3 runes left",
			s:     &Stream{source: bytes.NewReader([]byte("abc")), streamType: streamTypeText},
			r:     'a',
			size:  1,
			pos:   1,
			eos:   endOfStreamNot,
		},
		{
			title: "input text: 2 runes left",
			s:     &Stream{source: bytes.NewReader([]byte("bc")), streamType: streamTypeText, position: 1},
			r:     'b',
			size:  1,
			pos:   2,
			eos:   endOfStreamNot,
		},
		{
			title: "input text: 1 rune left, abrupt EOF",
			s:     &Stream{source: bytes.NewReader([]byte("c")), streamType: streamTypeText, position: 2},
			r:     'c',
			size:  1,
			pos:   3,
			eos:   endOfStreamNot,
		},
		{
			title: "input text: 1 rune left, non-abrupt EOF",
			s:     &Stream{source: newNonAbruptReader([]byte("c")), streamType: streamTypeText, position: 2},
			r:     'c',
			size:  1,
			pos:   3,
			eos:   endOfStreamAt,
		},
		{
			title: "input text: 1 rune left, file",
			s:     &Stream{source: mustOpen(testdata, "testdata/a.txt"), streamType: streamTypeText, position: 0},
			r:     'a',
			size:  1,
			pos:   1,
			eos:   endOfStreamAt,
		},
		{
			title: "input text: 1 rune left, file, failed to get file size",
			s:     &Stream{source: &m, streamType: streamTypeText, position: 0},
			r:     'a',
			size:  1,
			pos:   1,
			eos:   endOfStreamNot,
		},
		{
			title: "input Text: empty",
			s:     &Stream{source: bytes.NewReader([]byte("")), streamType: streamTypeText, position: 3},
			err:   io.EOF,
			pos:   3,
			eos:   endOfStreamPast,
		},
		{
			title: "end of stream past: error",
			s:     &Stream{source: bytes.NewReader([]byte("abc")), streamType: streamTypeText, endOfStream: endOfStreamPast, eofAction: eofActionError},
			err:   errPastEndOfStream,
			pos:   0,
			eos:   endOfStreamPast,
		},
		{
			title: "end of stream past: reset",
			s:     &Stream{source: bytes.NewReader([]byte("abc")), streamType: streamTypeText, endOfStream: endOfStreamPast, eofAction: eofActionReset, reposition: true},
			r:     'a',
			size:  1,
			pos:   1,
			eos:   endOfStreamNot,
		},
		{
			title: "input binary",
			s:     &Stream{source: bytes.NewReader([]byte("abc")), streamType: streamTypeBinary},
			err:   errWrongStreamType,
		},
		{
			title: "output",
			s:     &Stream{source: bytes.NewReader([]byte("abc")), mode: ioModeAppend},
			err:   errWrongIOMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			r, size, err := tt.s.ReadRune()
			equal(t, tt.r, r)
			equal(t, tt.size, size)
			equal(t, tt.err, err)

			equal(t, tt.pos, tt.s.position)
			equal(t, tt.eos, tt.s.endOfStream)
		})
	}
}

type mockSeeker struct {
	pos int64
	err error
}

func (m *mockSeeker) Seek(offset int64, whence int) (int64, error) {
	return m.pos, m.err
}

func TestStream_Seek(t *testing.T) {
	var okSeeker struct {
		mockReader
		mockWriter
		mockSeeker
	}

	var ngSeeker struct {
		mockReader
		mockWriter
		mockSeeker
	}
	ngSeeker.mockSeeker.err = errors.New("ng")

	s := &Stream{source: bytes.NewReader([]byte("abc")), streamType: streamTypeBinary, reposition: true}
	_, err := s.ReadByte()
	noError(t, err)

	tests := []struct {
		title  string
		s      *Stream
		offset int64
		whence int
		pos    int64
		err    error
	}{
		{
			title:  "ok input",
			s:      &Stream{source: &okSeeker, mode: ioModeRead, reposition: true},
			offset: 0,
			whence: 0,
		},
		{
			title:  "ok output",
			s:      &Stream{sink: &okSeeker, mode: ioModeWrite, reposition: true},
			offset: 0,
			whence: 0,
		},
		{
			title:  "ng input",
			s:      &Stream{source: &ngSeeker, mode: ioModeRead, reposition: true},
			offset: 0,
			whence: 0,
			err:    errors.New("ng"),
		},
		{
			title:  "ng output",
			s:      &Stream{sink: &ngSeeker, mode: ioModeWrite, reposition: true},
			offset: 0,
			whence: 0,
			err:    errors.New("ng"),
		},
		{title: "reader", s: s, offset: 0, whence: 0, pos: 0},
		{title: "reader", s: s, offset: 1, whence: 0, pos: 1},
		{title: "reader", s: s, offset: 2, whence: 0, pos: 2},
		{title: "reader", s: s, offset: 3, whence: 0, pos: 3},
		{
			title:  "not seeker",
			s:      &Stream{source: &okSeeker.mockReader, reposition: true, position: 123},
			offset: 0,
			whence: 0,
			pos:    123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			pos, err := tt.s.Seek(tt.offset, tt.whence)
			equal(t, tt.pos, pos)
			equal(t, tt.err, err)
		})
	}
}

func TestStream_WriteByte(t *testing.T) {
	var m mockWriter

	tests := []struct {
		title string
		s     *Stream
		c     byte
		err   error
		pos   int64
	}{
		{
			title: "writer",
			s:     &Stream{sink: &m, mode: ioModeAppend, streamType: streamTypeBinary},
			c:     byte('a'),
			pos:   1,
		},
		{
			title: "input",
			s:     &Stream{mode: ioModeRead, streamType: streamTypeBinary},
			c:     byte('a'),
			err:   errWrongIOMode,
			pos:   0,
		},
		{
			title: "text",
			s:     &Stream{mode: ioModeAppend, streamType: streamTypeText},
			c:     byte('a'),
			err:   errWrongStreamType,
			pos:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			err := tt.s.WriteByte(tt.c)
			equal(t, tt.err, err)

			equal(t, tt.pos, tt.s.position)
		})
	}
}

func TestStream_WriteRune(t *testing.T) {
	var m mockWriter

	tests := []struct {
		title string
		s     *Stream
		r     rune
		n     int
		err   error
		pos   int64
	}{
		{
			title: "writer",
			s:     &Stream{sink: &m, mode: ioModeAppend, streamType: streamTypeText},
			r:     'a',
			n:     1,
			pos:   1,
		},
		{
			title: "input",
			s:     &Stream{mode: ioModeRead, streamType: streamTypeText},
			r:     'a',
			err:   errWrongIOMode,
			pos:   0,
		},
		{
			title: "binary",
			s:     &Stream{mode: ioModeAppend, streamType: streamTypeBinary},
			r:     'a',
			err:   errWrongStreamType,
			pos:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			n, err := tt.s.WriteRune(tt.r)
			equal(t, tt.n, n)
			equal(t, tt.err, err)

			equal(t, tt.pos, tt.s.position)
		})
	}
}

type mockFlusher struct {
	err    error
	called bool
}

func (m *mockFlusher) Flush() error {
	m.called = true
	return m.err
}

type mockSyncer struct {
	err    error
	called bool
}

func (m *mockSyncer) Sync() error {
	m.called = true
	return m.err
}

func TestStream_Flush(t *testing.T) {
	t.Run("flusher", func(t *testing.T) {
		var m struct {
			mockWriter
			mockFlusher
		}

		s := &Stream{sink: &m, mode: ioModeAppend}
		noError(t, s.Flush())
		isTrue(t, m.called)
	})

	t.Run("syncer", func(t *testing.T) {
		var m struct {
			mockWriter
			mockSyncer
		}

		s := &Stream{sink: &m, mode: ioModeAppend}
		noError(t, s.Flush())
		isTrue(t, m.called)
	})

	t.Run("else", func(t *testing.T) {
		var m mockWriter

		s := &Stream{sink: &m, mode: ioModeAppend}
		noError(t, s.Flush())
	})
}

type nonAbruptReader struct {
	*bytes.Reader
}

func newNonAbruptReader(b []byte) nonAbruptReader {
	return nonAbruptReader{
		Reader: bytes.NewReader(b),
	}
}

func (r nonAbruptReader) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if err == nil && r.Len() == 0 {
		err = io.EOF
	}
	return n, err
}
