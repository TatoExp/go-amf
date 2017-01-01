package amf

import (
	"bytes"
	"encoding/hex"
	"io"
	"net"
	"reflect"
	"testing"
	"time"
)

// Encode

type encodeTestCase struct {
	in   interface{}
	want []byte
}

type encodeFunc func(interface{}) []byte

func testEncode(t *testing.T, cases []encodeTestCase, encode encodeFunc, name string) {
	for _, c := range cases {
		got := encode(c.in)
		if bytes.Compare(got, c.want) != 0 {
			t.Errorf("%s(%#v) == %#v (%d), want %#v (%d)", name, c.in, got, len(got), c.want, len(c.want))
		}
	}
}

// Decode

type decodeTestCase struct {
	in   []byte
	blen int
	want interface{}
}

type decodeFunc func([]byte) (interface{}, int, error)

func testDecode(t *testing.T, cases []decodeTestCase, decode decodeFunc, name string) {
	for _, c := range cases {
		got, blen, err := decode(c.in)
		if err != nil {
			t.Errorf("%s(%#v): %s", name, c.in, err)
			continue
		}
		if blen != c.blen || blen != len(c.in) {
			t.Errorf("%s(%#v) actual %d returned %d wanted %d", name, c.in, len(c.in), blen, c.blen)
			continue
		}
		switch got.(type) {
		case []interface{}, map[string]interface{}, ECMAArray:
			if !reflect.DeepEqual(c.want, got) {
				t.Errorf("%s(%#v) == %#v, want %#v", name, c.in, got, c.want)
			}
			continue
		}
		if got != c.want {
			t.Errorf("%s(%#v) == %#v, want %#v", name, c.in, got, c.want)
		}
	}
}

// Extern

func testExtern(t *testing.T, cases []decodeTestCase, name string) {
	conn, err := net.Dial("tcp", "localhost:4242")
	if err != nil {
		return
	}
	buffer := make([]byte, 128)
	for _, c := range cases {
		outdata := hex.EncodeToString(c.in)
		conn.SetDeadline(time.Now().Add(time.Second))
		t.Logf("sending [%s]", outdata)
		conn.Write(append([]byte(outdata), '\n'))
		ndata, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			t.Error(err)
			continue
		}
		indata := buffer[:ndata]
		t.Logf("received [%s]", string(indata))
		if outdata != string(indata) {
			t.Errorf("%s(%#v): %s != %s", name, c.in, indata, outdata)
		}
	}
	conn.Write([]byte("exit"))
	conn.Close()
}
