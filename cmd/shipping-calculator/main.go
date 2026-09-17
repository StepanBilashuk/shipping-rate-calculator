package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"shipping-rate-calculator/internal/shipping"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	return runWith(shipping.NewDefaultCalculator, args, stdout, stderr)
}

func runWith(newCalculator func() (*shipping.Calculator, error), args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("shipping-calculator", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		weight = fs.Float64("weight", 0, "parcel weight in kilograms")
		length = fs.Float64("length", 0, "parcel length in centimetres")
		width  = fs.Float64("width", 0, "parcel width in centimetres")
		height = fs.Float64("height", 0, "parcel height in centimetres")
		zone   = fs.String("zone", "", "destination zone: "+strings.Join(zoneNames(), ", "))
		asJSON = fs.Bool("json", false, "print the quote as JSON instead of text")
	)
	fs.Usage = func() { usage(stderr, fs) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", fs.Arg(0))
		usage(stderr, fs)
		return exitUsage
	}

	calc, err := newCalculator()
	if err != nil {
		printError(stderr, err)
		return exitFailure
	}

	quote, err := calc.Quote(parcelFromFlags(*weight, *length, *width, *height, *zone))
	if err != nil {
		printError(stderr, err)
		return exitFailure
	}

	if err := render(stdout, quote, *asJSON); err != nil {
		printError(stderr, err)
		return exitFailure
	}
	return exitOK
}

func printError(w io.Writer, err error) {
	fmt.Fprintf(w, "error: %s\n", strings.ReplaceAll(err.Error(), "\n", "\n       "))
}

func parcelFromFlags(weight, length, width, height float64, zone string) shipping.Parcel {
	return shipping.Parcel{
		WeightKg: weight,
		LengthCm: length,
		WidthCm:  width,
		HeightCm: height,
		Zone:     shipping.Zone(strings.ToLower(strings.TrimSpace(zone))),
	}
}

func render(w io.Writer, q shipping.Quote, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(q)
	}
	return renderText(w, q)
}

func renderText(w io.Writer, q shipping.Quote) error {
	labelWidth := len("Total")
	amountWidth := len(q.Total.String())
	for _, item := range q.Breakdown {
		labelWidth = max(labelWidth, len(item.Description))
		amountWidth = max(amountWidth, len(item.Amount.String()))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Parcel: %s\n", q.Parcel)
	for _, item := range q.Breakdown {
		fmt.Fprintf(&b, "  %-*s  %s %*s\n", labelWidth, item.Description, q.Currency, amountWidth, item.Amount)
	}
	fmt.Fprintf(&b, "  %s\n", strings.Repeat("-", labelWidth+amountWidth+len(q.Currency)+3))
	fmt.Fprintf(&b, "  %-*s  %s %*s\n", labelWidth, "Total", q.Currency, amountWidth, q.Total)

	_, err := io.WriteString(w, b.String())
	return err
}

func zoneNames() []string {
	zones := shipping.Zones()
	names := make([]string, len(zones))
	for i, z := range zones {
		names[i] = z.String()
	}
	return names
}

func usage(w io.Writer, fs *flag.FlagSet) {
	fmt.Fprint(w, `shipping-calculator calculates the shipping cost of a parcel, in EUR.

Usage:
  shipping-calculator -weight <kg> -length <cm> -width <cm> -height <cm> -zone <zone> [-json]

Flags:
`)
	fs.PrintDefaults()
	fmt.Fprintf(w, `
Example:
  shipping-calculator -weight 3 -length 20 -width 10 -height 5 -zone eu

Exit codes: %d success, %d invalid input, %d usage error.
`, exitOK, exitFailure, exitUsage)
}
