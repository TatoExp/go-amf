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
	case amf3Undefined:
		return nil, 1, nil
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
	case amf3Array:
		return decodeArray3(v)
	}
	return nil, 0, fmt.Errorf("unsupported type 0x%0X", v[0])
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
	return n, 1 + l, nil
}

func decodeDouble3(v []byte) (float64, int, error) {
	return math.Float64frombits(binary.BigEndian.Uint64(v[1:9])), 9, nil
}

func decodeUTF8VR(v []byte) (string, int, error) {
	strlen, l, err := decodeU29(v)
	if err != nil {
		return "", 0, err
	}
	if strlen&1 == 0 {
		return "", 0, fmt.Errorf("unsupported string ref")
	}
	strlen >>= 1
	return string(v[l : l+strlen]), l + strlen, nil
}

func decodeString3(v []byte) (string, int, error) {
	str, nstr, err := decodeUTF8VR(v[1:])
	if err != nil {
		return "", 0, err
	}
	return str, 1 + nstr, nil
}

func decodeDate3(v []byte) (time.Time, int, error) {
	if v[1] == 0 {
		return time.Time{}, 10, fmt.Errorf("unsupported date ref")
	}
	t := int64(math.Float64frombits(binary.BigEndian.Uint64(v[2:10])) * 1000000)
	return time.Unix(0, t), 10, nil
}

func decodeArray3(v []byte) (interface{}, int, error) {
	offset := 1
	num, nnum, err := decodeU29(v[offset:])
	if err != nil {
		return nil, 0, err
	}
	if num&1 == 0 {
		return nil, 0, fmt.Errorf("invalid array ref")
	}
	offset += nnum
	if num == 1 {
		return decodeAssociativeArray3(v, offset)
	} else {
		return decodeStrictArray3(v, offset, num>>1)
	}
}

func decodeAssociativeArray3(v []byte, offset int) (ECMAArray, int, error) {
	result := make(ECMAArray)
	for {
		key, nkey, err := decodeUTF8VR(v[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += nkey
		if key == "" {
			break
		}
		value, nvalue, err := decodeAMF3(v[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += nvalue
		result[key] = value
	}
	return result, offset, nil
}

func decodeStrictArray3(v []byte, offset, num int) ([]interface{}, int, error) {
	if v[offset] != 0x01 {
		return nil, 0, fmt.Errorf("invalid strict array")
	}
	offset++
	result := make([]interface{}, 0, num)
	for i := 0; i < num; i++ {
		value, nvalue, err := decodeAMF3(v[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += nvalue
		result = append(result, value)
	}
	return result, offset, nil
}
