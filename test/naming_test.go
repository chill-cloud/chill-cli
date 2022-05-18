package test

import (
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"testing"
)

func TestNaming(t *testing.T) {
	if naming.Validate("") {
		t.Fatal("empty name")
	}
	if !naming.Validate("coolstuff") {
		t.Fatal("regular name")
	}
	if !naming.Validate("cool-stuff") {
		t.Fatal("another regular name")
	}
	if naming.Validate("1coolname") {
		t.Fatal("starts with digit")
	}
	if !naming.Validate("coolname1") {
		t.Fatal("ends with digit")
	}
	if naming.Validate("coolname!") {
		t.Fatal("ends with exclamation mark")
	}
	if naming.Validate("cool--name") {
		t.Fatal("two minus in a row")
	}
	if naming.Validate("-coolname") {
		t.Fatal("starts with minus")
	}
	if naming.Validate("coolname-") {
		t.Fatal("ends with minus")
	}
	if naming.Validate("Coolname") {
		t.Fatal("has a capital letter")
	}
	if naming.Validate("COOLNAME") {
		t.Fatal("all capital letters")
	}
	if !naming.Validate("a") {
		t.Fatal("one-letter name is ok")
	}
}

func TestSplitting(t *testing.T) {
	s := "asd-qwe-zzz"
	if naming.MergeToCanonical(naming.SplitIntoParts(s)) != s {
		t.Fatal("not reversible")
	}
	if naming.Merge(naming.SplitIntoParts(s), "_", naming.ModeUpper) != "ASD_QWE_ZZZ" {
		t.Fatal("wrong upper underscore")
	}
	if naming.Merge(naming.SplitIntoParts(s), "_", naming.ModeLower) != "asd_qwe_zzz" {
		t.Fatal("wrong lower underscore")
	}
	if naming.Merge(naming.SplitIntoParts(s), "", naming.ModeUpperCamelCase) != "AsdQweZzz" {
		t.Fatal("wrong upper camel case")
	}
	if naming.Merge(naming.SplitIntoParts(s), "", naming.ModeLowerCamelCase) != "asdQweZzz" {
		t.Fatal("wrong lower camel case")
	}
}
