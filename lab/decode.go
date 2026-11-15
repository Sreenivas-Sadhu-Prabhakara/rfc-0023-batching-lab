package lab

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type Command struct {
	Offset int    `json:"offset"`
	Opcode int    `json:"opcode"`
	Name   string `json:"name"`
	Links  []int  `json:"links"`
}

type Report struct {
	Bytes    int       `json:"bytes"`
	Hex      string    `json:"hex"`
	Commands []Command `json:"commands"`
}

var opcodes = [...]struct {
	name string
	size int
}{
	{"Open primary", 1},
	{"Open auxiliary", 2},
	{"Accept primary", 2},
	{"Accept auxiliary", 2},
	{"Reject primary", 1},
	{"Reject auxiliary", 2},
	{"CEASE", 1},
	{"CLOSING", 1},
	{"UNCEASE", 1},
}

// DecodeHex parses whitespace-separated octets and decodes the full payload.
func DecodeHex(input string) (Report, error) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return Report{}, fmt.Errorf("empty payload")
	}
	payload := make([]byte, len(fields))
	for offset, field := range fields {
		if len(field) != 2 {
			return Report{}, fmt.Errorf("expected two hex digits at byte offset %d", offset)
		}
		if _, err := hex.Decode(payload[offset:offset+1], []byte(field)); err != nil {
			return Report{}, fmt.Errorf("invalid hex at byte offset %d: %w", offset, err)
		}
	}

	report := Report{Bytes: len(payload), Hex: fmt.Sprintf("% X", payload)}
	for offset := 0; offset < len(payload); {
		opcode := int(payload[offset])
		if opcode >= len(opcodes) {
			return Report{}, fmt.Errorf("unknown opcode 0x%02X at byte offset %d", opcode, offset)
		}
		spec := opcodes[opcode]
		end := offset + 1 + spec.size
		if end > len(payload) {
			return Report{}, fmt.Errorf("truncated opcode 0x%02X at byte offset %d: expected %d parameter bytes, got %d", opcode, offset, spec.size, len(payload)-offset-1)
		}
		command := Command{Offset: offset, Opcode: opcode, Name: spec.name, Links: make([]int, 0, spec.size)}
		for _, link := range payload[offset+1 : end] {
			command.Links = append(command.Links, int(link))
		}
		report.Commands = append(report.Commands, command)
		offset = end
	}
	return report, nil
}
