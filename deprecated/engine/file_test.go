package MTranCore

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
)

// Mock reader that implements readerWithLen interface
type mockReaderWithLen struct {
	data []byte
	pos  int
}

func (m *mockReaderWithLen) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockReaderWithLen) Len() int {
	return len(m.data) - m.pos
}

// Mock seeker reader
type mockSeeker struct {
	data []byte
	pos  int64
}

func (m *mockSeeker) Read(p []byte) (n int, err error) {
	if m.pos >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += int64(n)
	return n, nil
}

func (m *mockSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.pos = offset
	case io.SeekCurrent:
		m.pos += offset
	case io.SeekEnd:
		m.pos = int64(len(m.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if m.pos < 0 {
		m.pos = 0
	}
	if m.pos > int64(len(m.data)) {
		m.pos = int64(len(m.data))
	}

	return m.pos, nil
}

// Mock reader that fails on seek
type failingSeeker struct {
	data            []byte
	pos             int64
	failOnSeekEnd   bool
	failOnSeekStart bool
}

func (f *failingSeeker) Read(p []byte) (n int, err error) {
	if f.pos >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n = copy(p, f.data[f.pos:])
	f.pos += int64(n)
	return n, nil
}

func (f *failingSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekEnd && f.failOnSeekEnd {
		return 0, errors.New("seek end failed")
	}
	if whence == io.SeekStart && f.failOnSeekStart {
		return 0, errors.New("seek start failed")
	}

	switch whence {
	case io.SeekStart:
		f.pos = offset
	case io.SeekCurrent:
		f.pos += offset
	case io.SeekEnd:
		f.pos = int64(len(f.data)) + offset
	}

	return f.pos, nil
}

// Mock reader that fails on copy
type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read failed")
}

// plainReader is a simple reader that only implements io.Reader (not io.Seeker or readerWithLen)
type plainReader struct {
	data []byte
	pos  int
}

func newPlainReader(data string) *plainReader {
	return &plainReader{data: []byte(data)}
}

func (p *plainReader) Read(buf []byte) (n int, err error) {
	if p.pos >= len(p.data) {
		return 0, io.EOF
	}
	n = copy(buf, p.data[p.pos:])
	p.pos += n
	return n, nil
}

func TestAlignedMemoryFile_size(t *testing.T) {
	tests := []struct {
		name    string
		file    *alignedMemoryFile
		want    uint32
		wantErr bool
	}{
		{
			name: "nil reader",
			file: &alignedMemoryFile{
				Reader:    nil,
				Alignment: 4,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "bytes.Buffer reader",
			file: &alignedMemoryFile{
				Reader:    bytes.NewBufferString("hello world"),
				Alignment: 4,
			},
			want:    11,
			wantErr: false,
		},
		{
			name: "empty bytes.Buffer reader",
			file: &alignedMemoryFile{
				Reader:    bytes.NewBuffer(nil),
				Alignment: 4,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "readerWithLen implementation",
			file: &alignedMemoryFile{
				Reader:    &mockReaderWithLen{data: []byte("test data")},
				Alignment: 4,
			},
			want:    9,
			wantErr: false,
		},
		{
			name: "seeker reader",
			file: &alignedMemoryFile{
				Reader:    &mockSeeker{data: []byte("seeker test")},
				Alignment: 4,
			},
			want:    11,
			wantErr: false,
		},
		{
			name: "seeker reader - seek end fails",
			file: &alignedMemoryFile{
				Reader:    &failingSeeker{data: []byte("test"), failOnSeekEnd: true},
				Alignment: 4,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "seeker reader - seek start fails",
			file: &alignedMemoryFile{
				Reader:    &failingSeeker{data: []byte("test"), failOnSeekStart: true},
				Alignment: 4,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "generic reader",
			file: &alignedMemoryFile{
				Reader:    strings.NewReader("generic reader test"),
				Alignment: 4,
			},
			want:    19,
			wantErr: false,
		},
		{
			name: "generic reader - copy fails",
			file: &alignedMemoryFile{
				Reader:    &failingReader{},
				Alignment: 4,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "large file size",
			file: &alignedMemoryFile{
				Reader:    &mockReaderWithLen{data: make([]byte, 1000)},
				Alignment: 4,
			},
			want:    1000,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.file.size()
			if (err != nil) != tt.wantErr {
				t.Errorf("alignedMemoryFile.size() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("alignedMemoryFile.size() = %v, want %v", got, tt.want)
			}
		})
	}
}

// largeMockReader is a mock reader for testing large file sizes
type largeMockReader struct {
	size uint64
}

func (l *largeMockReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (l *largeMockReader) Len() int {
	return int(l.size)
}

func TestAlignedMemoryFile_size_TooLarge(t *testing.T) {
	// Test with a size that would exceed MaxUint32
	file := &alignedMemoryFile{
		Reader:    &largeMockReader{size: uint64(math.MaxUint32) + 1},
		Alignment: 4,
	}

	_, err := file.size()
	if err == nil {
		t.Error("expected error for file size exceeding MaxUint32")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' error message, got: %v", err)
	}
}

func TestAlignedMemoryFile_readAll(t *testing.T) {
	tests := []struct {
		name    string
		file    *alignedMemoryFile
		want    []byte
		wantErr bool
	}{
		{
			name: "nil reader",
			file: &alignedMemoryFile{
				Reader:    nil,
				Alignment: 4,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bytes.Buffer reader",
			file: &alignedMemoryFile{
				Reader:    bytes.NewBufferString("hello world"),
				Alignment: 4,
			},
			want:    []byte("hello world"),
			wantErr: false,
		},
		{
			name: "empty bytes.Buffer reader",
			file: &alignedMemoryFile{
				Reader:    bytes.NewBuffer(nil),
				Alignment: 4,
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "generic reader",
			file: &alignedMemoryFile{
				Reader:    strings.NewReader("generic reader test"),
				Alignment: 4,
			},
			want:    []byte("generic reader test"),
			wantErr: false,
		},
		{
			name: "generic reader - copy fails",
			file: &alignedMemoryFile{
				Reader:    &failingReader{},
				Alignment: 4,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.file.readAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("alignedMemoryFile.readAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("alignedMemoryFile.readAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlignedMemoryFile_readAll_ReaderReplacement(t *testing.T) {
	// Test that the reader is replaced with a buffer after reading from a generic reader
	originalReader := newPlainReader("test data")
	file := &alignedMemoryFile{
		Reader:    originalReader,
		Alignment: 4,
	}

	data, err := file.readAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(data, []byte("test data")) {
		t.Errorf("expected %q, got %q", "test data", string(data))
	}

	// Verify the reader was replaced with a buffer
	if _, ok := file.Reader.(*bytes.Buffer); !ok {
		t.Error("expected reader to be replaced with *bytes.Buffer")
	}
}

func TestAlignedMemoryFile_size_ReaderReplacement(t *testing.T) {
	// Test that the reader is replaced with a buffer after reading from a generic reader
	// Use plainReader which only implements io.Reader (not io.Seeker or readerWithLen)
	originalReader := newPlainReader("test data for size")
	file := &alignedMemoryFile{
		Reader:    originalReader,
		Alignment: 4,
	}

	size, err := file.size()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if size != 18 {
		t.Errorf("expected size 18, got %d", size)
	}

	// Verify the reader was replaced with a buffer
	if _, ok := file.Reader.(*bytes.Buffer); !ok {
		t.Error("expected reader to be replaced with *bytes.Buffer")
	}
}

func TestReaderWithLenInterface(t *testing.T) {
	// Test that bytes.Buffer implements readerWithLen
	var _ readerWithLen = (*bytes.Buffer)(nil)

	buf := bytes.NewBufferString("test")
	if buf.Len() != 4 {
		t.Errorf("expected buffer length 4, got %d", buf.Len())
	}
}

func TestReaderWithBytesInterface(t *testing.T) {
	// Test that bytes.Buffer implements readerWithBytes
	var _ readerWithBytes = (*bytes.Buffer)(nil)

	buf := bytes.NewBufferString("test")
	if !bytes.Equal(buf.Bytes(), []byte("test")) {
		t.Errorf("expected bytes %q, got %q", "test", string(buf.Bytes()))
	}
}
