package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	input, err := os.ReadFile("../../examples/burst.json")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(nil, bytes.NewReader(input), &output); err != nil || !strings.Contains(output.String(), `"last_ready": 12`) {
		t.Fatalf("JSON output %q, error %v", output.String(), err)
	}
	output.Reset()
	if err := run([]string{"-html"}, bytes.NewReader(input), &output); err != nil || !strings.Contains(output.String(), "Control message batching") {
		t.Fatalf("HTML output %q, error %v", output.String(), err)
	}
	for _, invalid := range []string{"{} {}", `{"unexpected":1}`, strings.Repeat(" ", (64<<10)+1)} {
		output.Reset()
		if err := run(nil, strings.NewReader(invalid), &output); err == nil || output.Len() != 0 {
			t.Fatalf("invalid input of %d bytes emitted %d bytes, error %v", len(invalid), output.Len(), err)
		}
	}
}
