package test

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func getFile() []byte {
	f, err := ioutil.ReadFile("../static/feed_example.xml")
	fmt.Print(err)
	return f
}

func TestPToStruct(t *testing.T) {

	p, err := ParseToStruct(getFile())
	if len(p.Building) == 0 || err != nil {
		t.Fatalf("cant parse %q cause of %v", getFile(), err)
	}
	t.Log("success")
}

func TestPToDB(t *testing.T) {
	err := ParseToDB(getFile())
	if err != nil {
		t.Fatalf("cant parse %q cause of %v. Res is ", getFile(), err)
	}
	t.Log("success")
}

func TestPFromDB(t *testing.T) {
	p, err := ConnectDB().Struct()
	if len(p.Building) == 0 || err != nil {
		t.Fatalf("cant get cause of %v", err)
	}
	t.Log("success")
}
