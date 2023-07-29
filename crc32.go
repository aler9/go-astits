package astits

const (
	crc32Polynomial = uint32(0xffffffff)
)

func computeCRC32(bs []byte) uint32 {
	return updateCRC32(crc32Polynomial, bs)
}

// Based on VLC implementation using a static CRC table (1kb additional memory on start, without
// reallocations): https://github.com/videolan/vlc/blob/master/modules/mux/mpeg/ps.c
func updateCRC32(crc32 uint32, bs []byte) uint32 {
	for _, b := range bs {
		crc32 = (crc32 << 8) ^ tableCRC32[((crc32>>24)^uint32(b))&0xff]
	}
	return crc32
}

type crc32Writer struct {
	w       *lightweightBitsWriter
	current uint32
}

func newCRC32Writer(w *lightweightBitsWriter) *crc32Writer {
	return &crc32Writer{
		w:       w,
		current: crc32Polynomial,
	}
}

func (c *crc32Writer) Write(p []byte) (int, error) {
	n := len(p)
	c.w.WriteSlice(p)
	err := c.w.Err()

	c.current = updateCRC32(c.current, p)

	return n, err
}

func (c *crc32Writer) Sum32() uint32 {
	return c.current
}
