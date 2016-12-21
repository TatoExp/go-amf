package amf

import (
	"bytes"
	"encoding/binary"
	"io"
	"sort"
	"time"
)

func EncodeAMF0(v interface{}) []byte {
	switch v.(type) {
	case float64:
		return encodeNumber(v.(float64))
	case int:
		return encodeNumber(float64(v.(int)))
	case bool:
		return encodeBoolean(v.(bool))
	case string:
		return encodeString(v.(string))
	case nil:
		return encodeNull()
	case map[string]interface{}:
		return encodeObject(v.(map[string]interface{}))
	case ECMAArray:
		return encodeECMAArray(v.(ECMAArray))
	case time.Time:
		return encodeDate(v.(time.Time))
	case []interface{}:
		return encodeStrictArray(v.([]interface{}))
	}
	return nil
}

func encodeNumber(v float64) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf0Number)
	binary.Write(buf, binary.BigEndian, v)
	return buf.Bytes()
}

func encodeBoolean(v bool) []byte {
	msg := make([]byte, 1+1) // 1 header + 1 boolean
	msg[0] = amf0Boolean
	if v {
		msg[1] = 0x1
	} else {
		msg[1] = 0x0
	}
	return msg
}

func encodeUTF8(w io.Writer, v string) {
	binary.Write(w, binary.BigEndian, uint16(len(v)))
	w.Write([]byte(v))
}

func encodeString(v string) []byte {
	buf := new(bytes.Buffer)
	if len(v) < 0xffff {
		buf.WriteByte(amf0String)
		encodeUTF8(buf, v)
	} else {
		buf.WriteByte(amf0StringExt)
		binary.Write(buf, binary.BigEndian, uint32(len(v)))
		buf.Write([]byte(v))
	}
	return buf.Bytes()
}

func encodeObject(v map[string]interface{}) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf0Object)
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := v[key]
		encodeUTF8(buf, key)
		buf.Write(EncodeAMF0(value))
	}
	encodeUTF8(buf, "")
	buf.WriteByte(amf0ObjectEnd)
	return buf.Bytes()
}

func encodeNull() []byte {
	return []byte{amf0Null}
}

func encodeECMAArray(v ECMAArray) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf0Array)
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	binary.Write(buf, binary.BigEndian, uint32(len(keys)))
	for _, key := range keys {
		value := v[key]
		encodeUTF8(buf, key)
		buf.Write(EncodeAMF0(value))
	}
	return buf.Bytes()
}

func encodeDate(v time.Time) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf0Date)
	binary.Write(buf, binary.BigEndian, float64(v.UnixNano()/1000000))
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)
	return buf.Bytes()
}

func encodeStrictArray(v []interface{}) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(amf0StrictArr)
	binary.Write(buf, binary.BigEndian, uint32(len(v)))
	for _, value := range v {
		buf.Write(EncodeAMF0(value))
	}
	return buf.Bytes()
}
