package amf

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

func DecodeAMF3(v []byte) (interface{}, error) {
	result, _, err := decodeAMF3(v)
	return result, err
}

func decodeAMF3(v []byte) (interface{}, int, error) {
	switch v[0] {
	case amf3Null:
		return nil, 1, nil
	case amf3False:
		return false, 1, nil
	case amf3True:
		return true, 1, nil
	case amf3Integer:
		return decodeInteger3(v)
	case amf3Double:
		return decodeDouble3(v)
	case amf3String:
		return decodeString3(v)
	case amf3Date:
		return decodeDate3(v)
	}
	return nil, 0, fmt.Errorf("unsupported type %X", v[0])
}

func decodeU29(v []byte) (int, int, error) {
	n := int(0)
	offset := 0
	for {
		if offset >= len(v) {
			return 0, 0, fmt.Errorf("EOF")
		}
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
	return n, offset, nil
}

func decodeInteger3(v []byte) (int, int, error) {
	n, l, err := decodeU29(v[1:])
	if err != nil {
		return 0, 0, err
	}
	if n&0x10000000 != 0 {
		n -= 0x20000000
	}
	return n, l, nil
}

func decodeDouble3(v []byte) (float64, int, error) {
	return math.Float64frombits(binary.BigEndian.Uint64(v[1:])), 9, nil
}

func decodeString3(v []byte) (string, int, error) {
	strlen, l, err := decodeU29(v[1:])
	if err != nil {
		return "", 0, err
	}
	return string(v[1+l:]), 1 + l + strlen, nil
}

func decodeDate3(v []byte) (time.Time, int, error) {
	if v[1] != 0x01 {
		return time.Time{}, 10, fmt.Errorf("invalid date tag")
	}
	return time.Unix(0, int64(binary.BigEndian.Uint64(v[2:10])*1000000)), 10, nil
}
