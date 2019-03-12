package main

import "io"

// An io.ReadSeeker that wraps another io.ReadSeeker and
// caches all of the data that is read.
type CachedReader struct {
	io.ReadSeeker
	Bitmap *Bitmap
	Data   []byte
}

func NewCachedReader(r io.ReadSeeker) (*CachedReader, error) {
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	return &CachedReader{
		ReadSeeker: r,
		Bitmap:     NewBitmap(int(size)),
		Data:       make([]byte, size),
	}, nil
}

func (c *CachedReader) Read(p []byte) (int, error) {
	offset, err := c.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	available := c.AvailableBytes(int(offset), len(p))
	if available > 0 {
		copy(p, c.Data[offset:])
		return available, nil
	}
	amount, err := c.ReadSeeker.Read(p)
	for i := 0; i < amount; i++ {
		c.Data[i+int(offset)] = p[i]
		c.Bitmap.Set(i + int(offset))
	}
	return amount, err
}

func (c *CachedReader) AvailableBytes(start, max int) int {
	for i := 0; i < max; i++ {
		if !c.Bitmap.Get(i + start) {
			return i
		}
	}
	return max
}

type Bitmap struct {
	Data    []uint8
	NumBits int
}

func NewBitmap(numBits int) *Bitmap {
	numBytes := numBits / 8
	if numBits&8 != 0 {
		numBytes += 1
	}
	return &Bitmap{
		Data:    make([]uint8, numBytes),
		NumBits: numBits,
	}
}

func (b *Bitmap) Get(i int) bool {
	return (b.Data[i/8] & (1 << uint(i&8))) != 0
}

func (b *Bitmap) Set(i int) {
	b.Data[i/8] |= (1 << uint(i&8))
}
