package amf

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"sort"
	"time"
)

const (
	amf3MaxInt = 268435455  // (2^28)-1
	amf3MinInt = -268435456 // -(2^28)
)

func EncodeAMF3(v interface{}) []byte {
	switch v.(type) {
	case float64:
		return encodeDouble3(v.(float64))
	case int:
		return encodeInteger3(v.(int))
	case uint:
		return encodeInteger3(v.(int))
	case bool:
		return encodeBoolean3(v.(bool))
	case string:
		return encodeString3(v.(string))
	case nil:
		return encodeNull3()
	// case map[string]interface{}:
		// return encodeObject3(v.(map[string]interface{}))
	case time.Time:
		return encodeDate3(v.(time.Time))
	case ECMAArray:
		return encodeAssociativeArray3(v.(ECMAArray))
	case []interface{}:
		return encodeStrictArray3(v.([]interface{}))
	}
	return nil
}

func encodeU29(w io.Writer, v int) {
	v &= 0x1fffffff
	if v <= 0x7f {
		w.Write([]byte{byte(v)})
	} else if v <= 0x3fff {
		w.Write([]byte{byte((v>>7)|0x80), byte(v&0x7f)})
	} else if v <= 0x1fffff {
		w.Write([]byte{byte((v>>14)|0x80), byte((v>>7)|0x80), byte(v&0x7f)})
	} else {
		w.Write([]byte{byte((v>>22)|0x80), byte((v>>14)|0x80), byte((v>>7)|0x80), byte(v)})
	}
}

func encodeInteger3(v int) []byte {
	if v >= amf3MinInt && v <= amf3MaxInt {
		buf := new(bytes.Buffer)
		buf.WriteByte(amf3Integer)
		encodeU29(buf, v)
		return buf.Bytes()
	} else {
		return encodeDouble3(float64(v))
	}
}

func encodeDouble3(v float64) []byte {
	msg := make([]byte, 1+8) // 1 header + 8 float64
	msg[0] = amf3Double
	binary.BigEndian.PutUint64(msg[1:], uint64(math.Float64bits(v)))
	return msg
}

func encodeBoolean3(v bool) []byte {
	if v {
		return []byte{amf3True}
	} else {
		return []byte{amf3False}
	}
}

func encodeNull3() []byte {
	return []byte{amf3Null}
}

func encodeUTF8VR(w io.Writer, v string) {
	var strlen = len(v)
	if strlen > amf3MaxInt {
		strlen = amf3MaxInt
	}
	encodeU29(w, (strlen<<1)|1)
	w.Write([]byte(v))
}

func encodeString3(v string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf3String)
	encodeUTF8VR(buf, v)
	return buf.Bytes()
}

func encodeDate3(v time.Time) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf3Date)
	encodeU29(buf, 1)
	binary.Write(buf, binary.BigEndian, float64(v.UnixNano()/1000000))
	return buf.Bytes()
}

func encodeAssociativeArray3(v ECMAArray) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf3Array)
	encodeU29(buf, 1)
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		b := EncodeAMF3(v[key])
		if b != nil {
			encodeUTF8VR(buf, key)
			buf.Write(b)
		}
	}
	buf.WriteByte(0x01)
	return buf.Bytes()
}

func encodeStrictArray3(v []interface{}) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf3Array)
	encodeU29(buf, (len(v)<<1)|1)
	buf.WriteByte(0x01)
	for _, item := range v {
		b := EncodeAMF3(item)
		if b != nil {
			buf.Write(b)
		}
	}
	return buf.Bytes()
}
