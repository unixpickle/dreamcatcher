package main

import (
	"io"
	"sync"
)

const MinRequestSize = 1 << 18

type DataCache struct {
	Lock   sync.RWMutex
	Bitmap *Bitmap
	Data   []byte
}

func NewDataCache(size int) *DataCache {
	return &DataCache{
		Bitmap: NewBitmap(size),
		Data:   make([]byte, size),
	}
}

func (d *DataCache) Available(offset, max int) int {
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	for i := 0; i < max; i++ {
		if i+offset >= len(d.Data) || !d.Bitmap.Get(i+offset) {
			return i
		}
	}
	return max
}

func (d *DataCache) GapSize(offset, max int) int {
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	for i := 0; i < max; i++ {
		if i+offset >= len(d.Data) || d.Bitmap.Get(i+offset) {
			return i
		}
	}
	return max
}

func (d *DataCache) WriteAt(offset int, data []byte) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	for i, b := range data {
		d.Bitmap.Set(i + offset)
		d.Data[i+offset] = b
	}
}

func (d *DataCache) ReadAt(offset int, p []byte) int {
	avail := d.Available(offset, len(p))
	if avail == 0 {
		return 0
	}
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	copy(p, d.Data[offset:offset+avail])
	return avail
}

// An io.ReadSeeker that wraps another io.ReadSeeker and
// caches all of the data that is read.
type CachedReader struct {
	io.ReadSeeker
	Cache *DataCache
}

func NewCachedReader(r io.ReadSeeker, cache *DataCache) *CachedReader {
	return &CachedReader{
		ReadSeeker: r,
		Cache:      cache,
	}
}

func (c *CachedReader) Read(p []byte) (int, error) {
	offset, err := c.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	if offset >= int64(len(c.Cache.Data)) {
		return 0, io.EOF
	}
	gapSize := c.Cache.GapSize(int(offset), MinRequestSize)
	if gapSize == 0 {
		n := c.Cache.ReadAt(int(offset), p)
		_, err = c.Seek(int64(n), io.SeekCurrent)
		return n, err
	}
	buffer := make([]byte, gapSize)
	amount, err := c.ReadSeeker.Read(buffer)
	c.Cache.WriteAt(int(offset), buffer[:amount])
	if amount > len(p) {
		_, err1 := c.Seek(int64(len(p)-amount), io.SeekCurrent)
		if err == nil {
			err = err1
		}
		amount = len(p)
	}
	copy(p, buffer[:amount])
	return amount, err
}

type Bitmap struct {
	Data    []uint8
	NumBits int
}

func NewBitmap(numBits int) *Bitmap {
	numBytes := numBits / 8
	if numBits&7 != 0 {
		numBytes += 1
	}
	return &Bitmap{
		Data:    make([]uint8, numBytes),
		NumBits: numBits,
	}
}

func (b *Bitmap) Get(i int) bool {
	return (b.Data[i>>3] & (1 << uint(i&7))) != 0
}

func (b *Bitmap) Set(i int) {
	b.Data[i>>3] |= 1 << uint(i&7)
}
