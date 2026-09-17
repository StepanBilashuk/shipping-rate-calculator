# Shipping Rate Calculator for Kuehne+Nagel.

Command-line tool for calculating parcel shipping costs in EUR.

Requires Go 1.26 or later. No third-party dependencies are used.

## Running it

```bash
go run ./cmd/shipping-calculator -weight 3 -length 20 -width 10 -height 5 -zone eu
```

```
Parcel: 3 kg, 20x10x5 cm, zone eu
  weight tier (over 1 kg up to 5 kg)  EUR 10.00
  zone surcharge (eu)                 EUR  8.00
  ---------------------------------------------
  Total                               EUR 18.00
```

Or build it first with `make build` and run `./bin/shipping-calculator`.

| Flag      | Meaning                                              |
|-----------|------------------------------------------------------|
| `-weight` | kilograms                                            |
| `-length` | centimetres                                          |
| `-width`  | centimetres                                          |
| `-height` | centimetres                                          |
| `-zone`   | `domestic`, `eu` or `international` (case-insensitive) |
| `-json`   | print JSON instead of the table above                |

All numeric values must be greater than zero. Dimensions are validated even
though the current tariff only uses weight.

Quotes are written to stdout and errors to stderr. Exit code `1` indicates
invalid input; `2` indicates incorrect command usage.

### JSON output

```json
{
  "parcel": { "weight": 3, "length": 20, "width": 10, "height": 5, "zone": "eu" },
  "breakdown": [
    { "rule": "weight_tier", "description": "weight tier (over 1 kg up to 5 kg)", "amount": "10.00" },
    { "rule": "zone_surcharge", "description": "zone surcharge (eu)", "amount": "8.00" }
  ],
  "total": "18.00",
  "currency": "EUR"
}
```

Amounts are represented as strings in JSON.

## The rules

| Weight        | Base rate | Zone          | Surcharge |
|---------------|-----------|---------------|-----------|
| up to 1 kg    | €5        | domestic      | +€0       |
| up to 5 kg    | €10       | eu            | +€8       |
| above 5 kg    | €20       | international | +€15      |

Both apply and the total is their sum. The bounds include their upper value, so
a parcel of exactly 5 kg still costs €10, not €20.

## Tests

```bash
make test                 # run all tests
make cover                # show coverage
make check                # formatting, vet, and tests with -race
go test ./... -short      # skip tests that build and run the binary
```

Tests cover tariff boundaries, all zones, invalid parcel data, malformed tariff
tables, JSON output, and command-line behaviour. The end-to-end test builds and
runs the binary.

## Implementation notes

The pricing code is in `internal/shipping`. A calculator applies the configured
rules and returns a breakdown together with the total. Weight tiers and zone
surcharges are kept as data, so rates can be changed without modifying the
calculation flow.

Amounts are stored as integer cents. The CLI is intentionally small: it parses
flags, creates a parcel, requests a quote, and formats the result.
