package lab

import (
	"fmt"
	"strings"
)

type Arrival struct {
	Tick int    `json:"tick"`
	Hex  string `json:"hex"`
}

type Scenario struct {
	AckDelay int       `json:"ack_delay"`
	MaxBytes int       `json:"max_bytes"`
	Arrivals []Arrival `json:"arrivals"`
}

type Transmission struct {
	Send  int `json:"send"`
	Ready int `json:"ready"`
	Report
}

type Run struct {
	Policy       string         `json:"policy"`
	LastReady    int            `json:"last_ready"`
	MaxQueueWait int            `json:"max_queue_wait"`
	Sends        []Transmission `json:"sends"`
}

type Comparison struct {
	AckDelay int   `json:"ack_delay"`
	MaxBytes int   `json:"max_bytes"`
	Commands int   `json:"commands"`
	Runs     []Run `json:"runs"`
}

// Compare drives both send policies with the same validated arrivals.
func Compare(s Scenario) (Comparison, error) {
	if s.AckDelay < 1 || s.AckDelay > 10000 {
		return Comparison{}, fmt.Errorf("ack_delay must be 1–10000 ticks")
	}
	if s.MaxBytes < 2 || s.MaxBytes > 1024 {
		return Comparison{}, fmt.Errorf("max_bytes must be 2–1024")
	}
	if len(s.Arrivals) < 1 || len(s.Arrivals) > 500 {
		return Comparison{}, fmt.Errorf("expected 1–500 arrivals")
	}

	commands := make([]Report, len(s.Arrivals))
	previous := 0
	for i, arrival := range s.Arrivals {
		if arrival.Tick < previous || arrival.Tick > 1000000 {
			return Comparison{}, fmt.Errorf("arrival %d: ticks must be nondecreasing and between 0 and 1000000", i+1)
		}
		previous = arrival.Tick
		command, err := DecodeHex(arrival.Hex)
		if err != nil {
			return Comparison{}, fmt.Errorf("arrival %d: %w", i+1, err)
		}
		if len(command.Commands) != 1 || command.Bytes > s.MaxBytes {
			return Comparison{}, fmt.Errorf("arrival %d: expected one complete command that fits max_bytes", i+1)
		}
		commands[i] = command
	}

	single, err := simulate(s, commands, false)
	if err != nil {
		return Comparison{}, err
	}
	batch, err := simulate(s, commands, true)
	if err != nil {
		return Comparison{}, err
	}
	return Comparison{AckDelay: s.AckDelay, MaxBytes: s.MaxBytes, Commands: len(commands), Runs: []Run{single, batch}}, nil
}

func simulate(s Scenario, commands []Report, batch bool) (Run, error) {
	result := Run{Policy: "One command per send"}
	if batch {
		result.Policy = "Batch on readiness"
	}
	for first := 0; first < len(commands); {
		sendTick := max(result.LastReady, s.Arrivals[first].Tick)
		end := first + 1
		size := commands[first].Bytes
		if batch {
			for end < len(commands) && s.Arrivals[end].Tick <= sendTick {
				if size+commands[end].Bytes > s.MaxBytes {
					break
				}
				size += commands[end].Bytes
				end++
			}
		}

		var payload strings.Builder
		for i := first; i < end; i++ {
			if i > first {
				payload.WriteByte(' ')
			}
			payload.WriteString(commands[i].Hex)
			result.MaxQueueWait = max(result.MaxQueueWait, sendTick-s.Arrivals[i].Tick)
		}
		decoded, err := DecodeHex(payload.String())
		if err != nil {
			return Run{}, fmt.Errorf("receive at tick %d: %w", sendTick, err)
		}
		result.LastReady = sendTick + s.AckDelay
		result.Sends = append(result.Sends, Transmission{Send: sendTick, Ready: result.LastReady, Report: decoded})
		first = end
	}
	return result, nil
}
