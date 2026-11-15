package lab

import (
	"reflect"
	"strings"
	"testing"
)

func TestBurstAndFIFO(t *testing.T) {
	s := Scenario{AckDelay: 4, MaxBytes: 7, Arrivals: []Arrival{
		{0, "00 12"}, {1, "02 12 34"}, {2, "06 34"}, {2, "08 34"}, {5, "07 34"},
	}}
	got, err := Compare(s)
	if err != nil {
		t.Fatal(err)
	}
	single, batch := got.Runs[0], got.Runs[1]
	if len(single.Sends) != 5 || single.LastReady != 20 || single.MaxQueueWait != 11 {
		t.Fatalf("unexpected baseline: %+v", single)
	}
	if len(batch.Sends) != 3 || batch.LastReady != 12 || batch.MaxQueueWait != 3 ||
		batch.Sends[0].Send != 0 || batch.Sends[1].Send != 4 ||
		batch.Sends[1].Hex != "02 12 34 06 34 08 34" || batch.Sends[2].Send != 8 {
		t.Fatalf("unexpected batch run: %+v", batch)
	}
	for _, run := range got.Runs {
		var actionOrder []int
		for _, send := range run.Sends {
			for _, command := range send.Commands {
				actionOrder = append(actionOrder, command.Opcode)
			}
		}
		if !reflect.DeepEqual(actionOrder, []int{0, 2, 6, 8, 7}) {
			t.Fatalf("%s: got opcode order %v", run.Policy, actionOrder)
		}
	}
}

func TestReadyBoundaryAndNoSkipping(t *testing.T) {
	s := Scenario{AckDelay: 3, MaxBytes: 4, Arrivals: []Arrival{
		{0, "00 10"}, {1, "06 10"}, {2, "02 10 20"}, {3, "08 10"}, {9, "07 10"}, {12, "06 11"},
	}}
	got, err := Compare(s)
	if err != nil {
		t.Fatal(err)
	}
	batch := got.Runs[1].Sends
	var ticks []int
	for _, send := range batch {
		ticks = append(ticks, send.Send)
		if send.Bytes > s.MaxBytes {
			t.Fatalf("payload exceeds cap: %+v", send)
		}
	}
	if !reflect.DeepEqual(ticks, []int{0, 3, 6, 9, 12}) ||
		batch[1].Hex != "06 10" || batch[3].Hex != "08 10 07 10" {
		t.Fatalf("unexpected batch grouping: %+v", batch)
	}
}

func TestInvalidArrivals(t *testing.T) {
	cases := []struct {
		arrivals []Arrival
		want     string
	}{
		{[]Arrival{{0, "00 12"}, {0, "01 34"}}, "arrival 2: truncated"},
		{[]Arrival{{0, "06 12 08 12"}}, "expected one complete command"},
		{[]Arrival{{0, "FF 12"}}, "unknown opcode"},
		{[]Arrival{{1, "00 12"}, {0, "06 12"}}, "nondecreasing"},
	}
	for _, tc := range cases {
		if _, err := Compare(Scenario{AckDelay: 4, MaxBytes: 7, Arrivals: tc.arrivals}); err == nil ||
			!strings.Contains(err.Error(), tc.want) {
			t.Errorf("got %v, want %q", err, tc.want)
		}
	}
	if _, err := Compare(Scenario{AckDelay: 1, MaxBytes: 2, Arrivals: []Arrival{{0, "02 12 34"}}}); err == nil {
		t.Fatal("command exceeds cap")
	}
}
