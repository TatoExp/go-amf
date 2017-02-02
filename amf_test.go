package amf

import (
	"bytes"
	"encoding/hex"
	"io"
	"net"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// Encode

type encodeTestCase struct {
	in   interface{}
	want []byte
}

type encodeFunc func(io.Writer, interface{}) (int, error)

func testEncode(t *testing.T, cases []encodeTestCase, encode encodeFunc, name string) {
	for _, c := range cases {
		buf := &bytes.Buffer{}
		n, err := encode(buf, c.in)
		if err != nil {
			t.Error(err)
		}
		if n != buf.Len() {
			t.Errorf("%s len got %d, want %d", name, n, buf.Len())
		}
		if bytes.Compare(buf.Bytes(), c.want) != 0 {
			t.Errorf("%s(%#v) == %#v (%d), want %#v (%d)", name, c.in, buf.Bytes(), buf.Len(), c.want, len(c.want))
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

func testExtern(t *testing.T, cases []decodeTestCase, name string, version int) {
	cmd := exec.Command("ruby", "amf_test_server.rb", strconv.Itoa(version))
	cmd.Start()

	conn, err := net.Dial("tcp", "localhost:4242")
	if err != nil {
		return
	}
	buffer := make([]byte, 256)
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
