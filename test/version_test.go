package test

import (
	"github.com/chill-cloud/chill-cli/pkg/version"
	"testing"
)

func TestVersions(t *testing.T) {
	res1, err := version.ParseFromString("v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	res2, err := version.ParseFromString("1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if res1.Compare(*res2) != 0 {
		t.Fatal("Equal versions are not parsed as equal")
	}
	res3, err := version.ParseFromString("v1")
	if err != nil {
		t.Fatal(err)
	}
	if res1.Compare(*res3) != 0 {
		t.Fatal("Equal versions are not parsed as equal")
	}
	res4, err := version.ParseFromString("v1.0")
	if err != nil {
		t.Fatal(err)
	}
	if res1.Compare(*res4) != 0 {
		t.Fatal("Equal versions are not parsed as equal")
	}
	_, err = version.ParseFromString("1.0.abc")
	if err == nil {
		t.Fatal("Wrong version parsed (trailing string)")
	}
	_, err = version.ParseFromString("1.0.0.0")
	if err == nil {
		t.Fatal("Wrong version parsed (too many parts)")
	}
	_, err = version.ParseFromString("-42.-111.-123")
	if err == nil {
		t.Fatal("Wrong version parsed (negative values)")
	}
}
