package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"shipping-rate-calculator/internal/shipping"
)

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut strings.Builder
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestRunTextOutput(t *testing.T) {
	code, stdout, stderr := runCLI(t, "-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu")

	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}
	want := strings.Join([]string{
		"Parcel: 3 kg, 20x10x5 cm, zone eu",
		"  weight tier (over 1 kg up to 5 kg)  EUR 10.00",
		"  zone surcharge (eu)                 EUR  8.00",
		"  ---------------------------------------------",
		"  Total                               EUR 18.00",
		"",
	}, "\n")
	if stdout != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
}

func TestRunJSONOutput(t *testing.T) {
	code, stdout, stderr := runCLI(t,
		"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu", "-json")

	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	var got struct {
		Total     string `json:"total"`
		Currency  string `json:"currency"`
		Breakdown []struct {
			Rule   string `json:"rule"`
			Amount string `json:"amount"`
		} `json:"breakdown"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if got.Total != "18.00" || got.Currency != "EUR" {
		t.Errorf("total = %s %s, want EUR 18.00", got.Currency, got.Total)
	}
	if len(got.Breakdown) != 2 {
		t.Fatalf("breakdown has %d items, want 2", len(got.Breakdown))
	}
	if got.Breakdown[0].Rule != "weight_tier" || got.Breakdown[1].Rule != "zone_surcharge" {
		t.Errorf("breakdown rules = %q, %q; want weight_tier, zone_surcharge",
			got.Breakdown[0].Rule, got.Breakdown[1].Rule)
	}
}

func TestRunAcceptsZoneInAnyCase(t *testing.T) {
	code, stdout, stderr := runCLI(t, "-weight", "1", "-length", "1", "-width", "1", "-height", "1", "-zone", "EU")

	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}
	if !strings.Contains(stdout, "zone eu") {
		t.Errorf("stdout =\n%s\nwant it to report zone eu", stdout)
	}
}

func TestRunErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantErr  string
	}{
		{
			name:     "missing zone",
			args:     []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5"},
			wantCode: exitFailure,
			wantErr:  "zone: required",
		},
		{
			name:     "unknown zone",
			args:     []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "moon"},
			wantCode: exitFailure,
			wantErr:  `zone: "moon" is not one of`,
		},
		{
			name:     "missing dimensions",
			args:     []string{"-weight", "3", "-zone", "eu"},
			wantCode: exitFailure,
			wantErr:  "length",
		},
		{
			name:     "negative weight",
			args:     []string{"-weight", "-3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu"},
			wantCode: exitFailure,
			wantErr:  "weight",
		},
		{
			name:     "unknown flag",
			args:     []string{"-nope"},
			wantCode: exitUsage,
			wantErr:  "not defined",
		},
		{
			name:     "unexpected positional argument",
			args:     []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu", "extra"},
			wantCode: exitUsage,
			wantErr:  "unexpected argument",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(t, tt.args...)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %s)", code, tt.wantCode, stderr)
			}
			if !strings.Contains(stderr, tt.wantErr) {
				t.Errorf("stderr = %q, want it to mention %q", stderr, tt.wantErr)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want nothing on failure", stdout)
			}
		})
	}
}

func TestRunReportsEveryProblemAtOnce(t *testing.T) {
	code, _, stderr := runCLI(t, "-weight", "0", "-length", "-2", "-width", "10", "-height", "5", "-zone", "moon")

	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}
	for _, want := range []string{"weight", "length", "zone"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr = %q, want it to mention %q", stderr, want)
		}
	}
}

func TestRunHelp(t *testing.T) {
	code, stdout, stderr := runCLI(t, "-h")

	if code != exitOK {
		t.Errorf("exit code = %d, want %d", code, exitOK)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want the usage text on stderr", stdout)
	}
	for _, want := range []string{"Usage:", "-weight", "-zone", "Example:"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("usage text does not mention %q:\n%s", want, stderr)
		}
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestRunReportsOutputFailure(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
	}{
		{"text output", []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu"}},
		{"json output", []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu", "-json"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stderr strings.Builder
			broken := failingWriter{err: errors.New("disk on fire")}

			code := run(tt.args, broken, &stderr)

			if code != exitFailure {
				t.Errorf("exit code = %d, want %d", code, exitFailure)
			}
			if !strings.Contains(stderr.String(), "disk on fire") {
				t.Errorf("stderr = %q, want it to report the write failure", stderr.String())
			}
		})
	}
}

func TestRunReportsCalculatorFailure(t *testing.T) {
	var stdout, stderr strings.Builder
	broken := func() (*shipping.Calculator, error) {
		return nil, errors.New("tariff table rejected")
	}

	code := runWith(broken, []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu"},
		&stdout, &stderr)

	if code != exitFailure {
		t.Errorf("exit code = %d, want %d", code, exitFailure)
	}
	if !strings.Contains(stderr.String(), "tariff table rejected") {
		t.Errorf("stderr = %q, want it to report the rejected tariff", stderr.String())
	}
	if stdout.String() != "" {
		t.Errorf("stdout = %q, want nothing", stdout.String())
	}
}
