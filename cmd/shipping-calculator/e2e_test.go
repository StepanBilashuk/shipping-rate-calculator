package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	dir, err := os.MkdirTemp("", "shipping-calculator-e2e")
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating a temporary directory: %v\n", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(dir, "shipping-calculator")
	if out, err := exec.Command("go", "build", "-o", binaryPath, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building the CLI: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("-short: skipping the tests that build and run the binary")
	}

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:     "prices the parcel from the specification",
			args:     []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "eu"},
			wantCode: exitOK,
			wantStdout: strings.Join([]string{
				"Parcel: 3 kg, 20x10x5 cm, zone eu",
				"  weight tier (over 1 kg up to 5 kg)  EUR 10.00",
				"  zone surcharge (eu)                 EUR  8.00",
				"  ---------------------------------------------",
				"  Total                               EUR 18.00",
				"",
			}, "\n"),
		},
		{
			name:       "invalid input exits 1 and writes to stderr",
			args:       []string{"-weight", "3", "-length", "20", "-width", "10", "-height", "5", "-zone", "moon"},
			wantCode:   exitFailure,
			wantStderr: `zone: "moon" is not one of`,
		},
		{
			name:       "a usage error exits 2",
			args:       []string{"-nope"},
			wantCode:   exitUsage,
			wantStderr: "not defined",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			_ = cmd.Run()
			code := cmd.ProcessState.ExitCode()

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %s)", code, tt.wantCode, stderr.String())
			}
			if tt.wantStdout != "" && stdout.String() != tt.wantStdout {
				t.Errorf("stdout =\n%s\nwant\n%s", stdout.String(), tt.wantStdout)
			}
			if tt.wantStdout == "" && stdout.String() != "" {
				t.Errorf("stdout = %q, want nothing on failure", stdout.String())
			}
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to mention %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}
