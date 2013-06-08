package iproto

import (
	"fmt"
	"testing"
	"bytes"
)

func TestConnect(t *testing.T) {
	conn, err := Connect("localhost:33013")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	var rid int32 = 17
	resp, err := conn.Request(rid, new(bytes.Buffer))
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}