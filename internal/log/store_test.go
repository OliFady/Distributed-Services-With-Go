package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	write = []byte("Hello World")
	width = uint64(len(write)) + lenWidth
)

func TestStoreAppendRead(t *testing.T) {
	f, err := os.CreateTemp("", "store_append_read_test") //creates a temporary file
	require.NoError(t, err)
	defer os.Remove(f.Name()) //cleans up

	s, err := newStore(f)
	require.NoError(t, err)

	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)

	s, err = newStore(f)
	require.NoError(t, err)
	testRead(t, s)
}

func testAppend(t *testing.T, s *store) {
	t.Helper() //helps with testing

	for i := uint64(1); i < 4; i++ {
		n, pos, err := s.Append(write)
		require.NoError(t, err)
		require.Equal(t, i*width, pos+n)
	}
}

func testRead(t *testing.T, s *store) {
	t.Helper() //helps with testing

	var pos uint64
	for i := uint64(1); i < 4; i++ {
		read, err := s.Read(pos)
		require.NoError(t, err)
		require.Equal(t, write, read)
		pos += width
	}
}

func testReadAt(t *testing.T, s *store) {
	t.Helper() //helps with testing

	for i, off := uint64(1), int64(0); i < 4; i++ {
		b := make([]byte, lenWidth)
		n, err := s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, lenWidth, n)
		off += int64(n)

		size := enc.Uint64(b)
		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, size, b)
		require.Equal(t, int(size), n)
		off += int64(n)
	}
}

func TestStoreClose(t *testing.T) {
	f, err := os.CreateTemp("", "store_close_test") //creates a temporary file
	require.NoError(t, err)
	defer os.Remove(f.Name()) //cleans up
	s, err := newStore(f)
	require.NoError(t, err)
	_, _, err = s.Append(write)
	require.NoError(t, err)

	f, beforeSize, err := openFile(f.Name())
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, afterSize, err := openFile(f.Name())
	require.NoError(t, err)
	require.True(t, beforeSize < afterSize)
}

func openFile(name string) (file *os.File, size int64, err error) {
	file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644) //opens the file
	if err != nil {
		return nil, 0, err
	}
	fi, err := file.Stat() //gets the size of the file
	if err != nil {
		return nil, 0, err
	}
	return file, fi.Size(), nil
}
