package amf

import (
	"encoding/binary"
	"math"
)

func DecodeAMF3(v []byte) interface{} {
	switch v[0] {
	case amf3Null:
		return nil
	case amf3False:
		return false
	case amf3True:
		return true
	case amf3Integer:
		return decodeInteger3(v)
	case amf3Double:
		return decodeDouble3(v)
	}
	return nil
}

func decodeU29(v []byte) int {
	n := int(0)
	offset := 0
	for {
		b := int(v[offset])
		offset++
		if offset == 4 {
			n <<= 8
			n |= b
			break
		}
		n <<= 7
		n |= b & 0x7f
		if b&0x80 == 0 {
			break
		}
	}
	return n
}

func decodeInteger3(v []byte) int {
	n := decodeU29(v[1:])
	if n&0x10000000 != 0 {
		n -= 0x20000000
	}
	return n
}

func decodeDouble3(v []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(v[1:]))
}
