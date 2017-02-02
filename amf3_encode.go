package amf

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
	"time"
)

const (
	amf3MaxInt = 268435455  // (2^28)-1
	amf3MinInt = -268435456 // -(2^28)
)

func EncodeAMF3(w io.Writer, v interface{}) (int, error) {
	return encodeAMF3(0, w, v)
}

func encodeAMF3(n int, w io.Writer, v interface{}) (int, error) {
	switch v.(type) {
	case float64:
		return encodeDouble3(n, w, v.(float64))
	case int:
		return encodeInteger3(n, w, v.(int))
	case uint:
		return encodeInteger3(n, w, v.(int))
	case bool:
		return encodeBoolean3(n, w, v.(bool))
	case string:
		return encodeString3(n, w, v.(string))
	case nil:
		return encodeNull3(n, w)
	// case map[string]interface{}:
	// return encodeObject3(n, w, v.(map[string]interface{}))
	case time.Time:
		return encodeDate3(n, w, v.(time.Time))
	case ECMAArray:
		return encodeAssociativeArray3(n, w, v.(ECMAArray))
	case []interface{}:
		return encodeStrictArray3(n, w, v.([]interface{}))
	}
	return n, fmt.Errorf("type %T not supported", v)
}

func encodeU29(n int, w io.Writer, v int) (int, error) {
	v &= 0x1fffffff
	if v <= 0x7f {
		return writeBytes(n, w, []byte{byte(v)})
	} else if v <= 0x3fff {
		return writeBytes(n, w, []byte{byte((v >> 7) | 0x80), byte(v & 0x7f)})
	} else if v <= 0x1fffff {
		return writeBytes(n, w, []byte{byte((v >> 14) | 0x80), byte((v >> 7) | 0x80), byte(v & 0x7f)})
	} else {
		return writeBytes(n, w, []byte{byte((v >> 22) | 0x80), byte((v >> 14) | 0x80), byte((v >> 7) | 0x80), byte(v)})
	}
}

func encodeInteger3(n int, w io.Writer, v int) (int, error) {
	if v >= amf3MinInt && v <= amf3MaxInt {
		n, err := writeBytes(n, w, []byte{amf3Integer})
		if err != nil {
			return n, err
		}
		return encodeU29(n, w, v)
	} else {
		return encodeDouble3(n, w, float64(v))
	}
}

func encodeDouble3(n int, w io.Writer, v float64) (int, error) {
	n, err := writeBytes(n, w, []byte{amf3Double})
	if err != nil {
		return n, err
	}
	return writeData(n, w, binary.BigEndian, uint64(math.Float64bits(v)))
}

func encodeBoolean3(n int, w io.Writer, v bool) (int, error) {
	if v {
		return writeBytes(n, w, []byte{amf3True})
	} else {
		return writeBytes(n, w, []byte{amf3False})
	}
}

func encodeNull3(n int, w io.Writer) (int, error) {
	return writeBytes(n, w, []byte{amf3Null})
}

func encodeUTF8VR(n int, w io.Writer, v string) (int, error) {
	var strlen = len(v)
	if strlen > amf3MaxInt {
		strlen = amf3MaxInt
	}
	n, err := encodeU29(n, w, (strlen<<1)|1)
	if err != nil {
		return n, err
	}
	return writeBytes(n, w, []byte(v))
}

func encodeString3(n int, w io.Writer, v string) (int, error) {
	n, err := writeBytes(n, w, []byte{amf3String})
	if err != nil {
		return n, err
	}
	return encodeUTF8VR(n, w, v)
}

func encodeDate3(n int, w io.Writer, v time.Time) (int, error) {
	n, err := writeBytes(n, w, []byte{amf3Date})
	if err != nil {
		return n, err
	}
	n, err = encodeU29(n, w, 1)
	if err != nil {
		return n, err
	}
	return writeData(n, w, binary.BigEndian, float64(v.UnixNano()/1000000))
}

func encodeAssociativeArray3(n int, w io.Writer, v ECMAArray) (int, error) {
	n, err := writeBytes(n, w, []byte{amf3Array})
	if err != nil {
		return n, err
	}
	n, err = encodeU29(n, w, 1)
	if err != nil {
		return n, err
	}
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		n, err = encodeUTF8VR(n, w, key)
		if err != nil {
			return n, err
		}
		n, err = encodeAMF3(n, w, v[key])
		if err != nil {
			return n, err
		}
	}
	return writeBytes(n, w, []byte{0x01})
}

func encodeStrictArray3(n int, w io.Writer, v []interface{}) (int, error) {
	n, err := writeBytes(n, w, []byte{amf3Array})
	if err != nil {
		return n, err
	}
	n, err = encodeU29(n, w, (len(v)<<1)|1)
	if err != nil {
		return n, err
	}
	n, err = writeBytes(n, w, []byte{0x01})
	if err != nil {
		return n, err
	}
	for _, item := range v {
		n, err = encodeAMF3(n, w, item)
		if err != nil {
			return n, err
		}
	}
	return n, err
}
