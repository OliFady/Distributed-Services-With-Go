package log

import (
	"io"
	"os"

	"launchpad.net/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		//memory maps the file
		return nil, err
	}
	return idx, nil
}

func (idx *index) Close() error {
	if err := idx.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := idx.file.Sync(); err != nil {
		return err
	}
	if err := idx.file.Truncate(int64(idx.size)); err != nil {
		return err
	}
	return idx.file.Close()
}

func (idx *index) Read(in int64) (out uint32, pos uint64, err error) {

	if idx.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((idx.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if idx.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(idx.mmap[pos : pos+offWidth])
	pos = enc.Uint64(idx.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (idx *index) Write(off uint32, pos uint64) error {
	if uint64(len(idx.mmap)) < idx.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(idx.mmap[idx.size:idx.size+offWidth], off)          //writes the offset to the file
	enc.PutUint64(idx.mmap[idx.size+offWidth:idx.size+entWidth], pos) //writes the position to the file
	idx.size += uint64(entWidth)
	return nil
}

func (idx *index) Name() string {
	return idx.file.Name()
}
