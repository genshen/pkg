package pkg

import (
	"testing"
)

func TestParseInsCmdOnly(t *testing.T) {
	if ins, err := ParseIns(`CP`); err != nil || ins.First != "CP" {
		t.Error("test error of cmd only")
	}
	if ins, err := ParseIns(` CP`); err != nil || ins.First != "CP" {
		t.Error("test error of cmd only")
	}
	if ins, err := ParseIns(`CP `); err != nil || ins.First != "CP" {
		t.Error("test error of cmd only")
	}
	if ins, err := ParseIns(` CP `); err != nil || ins.First != "CP" {
		t.Error("test error of cmd only")
	}
}

func TestParseInsSecond(t *testing.T) {
	if ins, err := ParseIns(`CP ../path ./`); err != nil {
		t.Error("test error of ins second parsing")
	} else {
		if ins.Second != "../path" || ins.Third != "./" {
			t.Errorf("test error of ins second parsing: `CP ../path ./`, second:%s, third:%s",
				ins.Second, ins.Third)
		}
	}

	if ins, err := ParseIns(`CP ../path `); err != nil {
		t.Error("test error of ins second parsing")
	} else {
		if ins.Second != "../path" || ins.Third != "" {
			t.Errorf("test error of ins second parsing: `CP ../path `, second:%s, third:%s",
				ins.Second, ins.Third)
		}
	}

	if ins, err := ParseIns(`CP ../path "./des" `); err != nil {
		t.Error("test error of ins second parsing")
	} else {
		if ins.Second != "../path" || ins.Third != "./des" {
			t.Errorf("test error of ins second parsing: `CP ../path \"./des\"`, second:%s, third:%s",
				ins.Second, ins.Third)
		}
	}
}

func TestParseInsSecondQ(t *testing.T) {
	if _, err := ParseIns(`CP "`); err == nil {
		t.Error("test error of ins second parsing")
	}
	if _, err := ParseIns(`CP " `); err == nil {
		t.Error("test error of ins second parsing")
	}

	if ins, err := ParseIns(`CP ""`); err != nil {
		t.Error("test error of ins second parsing")
	} else {
		if ins.Second != "" {
			t.Errorf("test error of ins second parsing, second:%s", ins.Second)
		}
	}

	if ins, err := ParseIns(`CP "C"`); err != nil || ins.Second != "C" {
		t.Error("test error of ins second parsing")
	}
	if ins, err := ParseIns("CP \"C \""); err != nil || ins.Second != "C" {
		t.Error("test error of ins second parsing")
	}
}
