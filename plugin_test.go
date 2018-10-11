package main

import "testing"

func TestPluginTruth(t *testing.T) {
	if true != true {
		t.Error("everything I know is wrong")
	}
}
