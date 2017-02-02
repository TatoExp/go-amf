package amf

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"time"
)

func EncodeAMF0(w io.Writer, v interface{}) (int, error) {
	return encodeAMF0(0, w, v)
}

func encodeAMF0(n int, w io.Writer, v interface{}) (int, error) {
	switch v.(type) {
	case float64:
		return encodeNumber(n, w, v.(float64))
	case int:
		return encodeNumber(n, w, float64(v.(int)))
	case bool:
		return encodeBoolean(n, w, v.(bool))
	case string:
		return encodeString(n, w, v.(string))
	case nil:
		return encodeNull(n, w)
	case map[string]interface{}:
		return encodeObject(n, w, v.(map[string]interface{}))
	case ECMAArray:
		return encodeECMAArray(n, w, v.(ECMAArray))
	case time.Time:
		return encodeDate(n, w, v.(time.Time))
	case []interface{}:
		return encodeStrictArray(n, w, v.([]interface{}))
	}
	return n, fmt.Errorf("type %T not supported", v)
}

func encodeNumber(n int, w io.Writer, v float64) (int, error) {
	n, err := writeBytes(n, w, []byte{amf0Number})
	if err != nil {
		return n, err
	}
	return writeData(n, w, binary.BigEndian, v)
}

func encodeBoolean(n int, w io.Writer, v bool) (int, error) {
	if v {
		return writeBytes(n, w, []byte{amf0Boolean, 0x1})
	} else {
		return writeBytes(n, w, []byte{amf0Boolean, 0x0})
	}
}

func encodeUTF8(n int, w io.Writer, v string) (int, error) {
	var err error
	n, err = writeData(n, w, binary.BigEndian, uint16(len(v)))
	if err != nil {
		return n, err
	}
	return writeBytes(n, w, []byte(v))
}

func encodeString(n int, w io.Writer, v string) (int, error) {
	if len(v) < 0xffff {
		n, err := writeBytes(n, w, []byte{amf0String})
		if err != nil {
			return n, err
		}
		return encodeUTF8(n, w, v)
	} else {
		n, err := writeBytes(n, w, []byte{amf0StringExt})
		if err != nil {
			return n, err
		}
		n, err = writeData(n, w, binary.BigEndian, uint32(len(v)))
		if err != nil {
			return n, err
		}
		return writeBytes(n, w, []byte(v))
	}
}

func encodeObject(n int, w io.Writer, v map[string]interface{}) (int, error) {
	n, err := writeBytes(n, w, []byte{amf0Object})
	if err != nil {
		return n, err
	}
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := v[key]
		n, err = encodeUTF8(n, w, key)
		if err != nil {
			return n, err
		}
		n, err = encodeAMF0(n, w, value)
		if err != nil {
			return n, err
		}
	}
	n, err = encodeUTF8(n, w, "")
	if err != nil {
		return n, err
	}
	return writeBytes(n, w, []byte{amf0ObjectEnd})
}

func encodeNull(n int, w io.Writer) (int, error) {
	return writeBytes(n, w, []byte{amf0Null})
}

func encodeECMAArray(n int, w io.Writer, v ECMAArray) (int, error) {
	n, err := writeBytes(n, w, []byte{amf0Array})
	if err != nil {
		return n, err
	}
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	n, err = writeData(n, w, binary.BigEndian, uint32(len(keys)))
	if err != nil {
		return n, err
	}
	for _, key := range keys {
		value := v[key]
		n, err = encodeUTF8(n, w, key)
		if err != nil {
			return n, err
		}
		n, err = encodeAMF0(n, w, value)
		if err != nil {
			return n, err
		}
	}
	n, err = encodeUTF8(n, w, "")
	if err != nil {
		return n, err
	}
	return writeBytes(n, w, []byte{amf0ObjectEnd})
}

func encodeDate(n int, w io.Writer, v time.Time) (int, error) {
	n, err := writeBytes(n, w, []byte{amf0Date})
	if err != nil {
		return n, err
	}
	n, err = writeData(n, w, binary.BigEndian, float64(v.UnixNano()/1000000))
	if err != nil {
		return n, err
	}
	return writeBytes(n, w, []byte{0x00, 0x00})
}

func encodeStrictArray(n int, w io.Writer, v []interface{}) (int, error) {
	n, err := writeBytes(n, w, []byte{amf0StrictArr})
	if err != nil {
		return n, err
	}
	n, err = writeData(n, w, binary.BigEndian, uint32(len(v)))
	if err != nil {
		return n, err
	}
	for _, value := range v {
		n, err = encodeAMF0(n, w, value)
		if err != nil {
			return n, err
		}
	}
	return n, err
}
