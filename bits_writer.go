package astits

import (
	"io"
)

// lightweight version of astikit.BitWriter, without
// - dynamic endianness
// - WriteCallback
// - generic Write()
// - errors returned after each write
type lightweightBitsWriter struct {
	w        io.Writer
	cache    byte
	cacheLen byte
	bsCache  []byte
	err      error
}

func newLightweightBitsWriter(w io.Writer) *lightweightBitsWriter {
	return &lightweightBitsWriter{
		w:       w,
		bsCache: make([]byte, 1),
	}
}

func (w *lightweightBitsWriter) Err() error {
	return w.err
}

func (w *lightweightBitsWriter) flushBsCache() {
	if w.err == nil {
		_, w.err = w.w.Write(w.bsCache)
	}
}

func (w *lightweightBitsWriter) WriteBit(bit bool) {
	if bit {
		w.cache |= 1 << (7 - w.cacheLen)
	}
	w.cacheLen++
	if w.cacheLen == 8 {
		w.bsCache[0] = w.cache
		w.flushBsCache()
		w.cacheLen = 0
		w.cache = 0
	}
}

func (w *lightweightBitsWriter) WriteBits(toWrite uint64, n int) {
	toWrite &= ^uint64(0) >> (64 - n)

	for n > 0 {
		if w.cacheLen == 0 {
			if n >= 8 {
				n -= 8
				w.bsCache[0] = byte(toWrite >> n)
				w.flushBsCache()
			} else {
				w.cacheLen = uint8(n)
				w.cache = byte(toWrite << (8 - w.cacheLen))
				n = 0
			}
		} else {
			free := int(8 - w.cacheLen)
			m := n
			if m >= free {
				m = free
			}

			if n <= free {
				w.cache |= byte(toWrite << (free - m))
			} else {
				w.cache |= byte(toWrite >> (n - m))
			}

			n -= m
			w.cacheLen += uint8(m)

			if w.cacheLen == 8 {
				w.bsCache[0] = w.cache
				w.flushBsCache()

				w.cacheLen = 0
				w.cache = 0
			}
		}
	}
}

func (w *lightweightBitsWriter) WriteByte(b uint8) {
	if w.cacheLen == 0 {
		w.bsCache[0] = b
	} else {
		w.bsCache[0] = w.cache | (b >> w.cacheLen)
		w.cache = b << (8 - w.cacheLen)
	}
	w.flushBsCache()
}

func (w *lightweightBitsWriter) WriteUint16(v uint16) {
	w.WriteBits(uint64(v), 16)
}

func (w *lightweightBitsWriter) WriteUint32(v uint32) {
	w.WriteBits(uint64(v), 32)
}

func (w *lightweightBitsWriter) WriteSlice(in []byte) {
	if len(in) == 0 {
		return
	}

	if w.cacheLen != 0 {
		for _, b := range in {
			w.WriteByte(b)
		}
	} else {
		if w.err == nil {
			_, w.err = w.w.Write(in)
		}
	}
}
