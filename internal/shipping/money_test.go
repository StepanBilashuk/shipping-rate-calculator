package shipping

import (
	"encoding/json"
	"testing"
)

func TestMoneyString(t *testing.T) {
	tests := []struct {
		name  string
		money Money
		want  string
	}{
		{"zero", Euros(0), "0.00"},
		{"whole euros", Euros(18), "18.00"},
		{"euros and cents", Cents(1805), "18.05"},
		{"cents below ten keep the leading zero", Cents(105), "1.05"},
		{"cents only", Cents(7), "0.07"},
		{"negative amounts keep the sign", Cents(-250), "-2.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.money.String(); got != tt.want {
				t.Errorf("Money(%d).String() = %q, want %q", int64(tt.money), got, tt.want)
			}
		})
	}
}

func TestMoneyAdditionIsExact(t *testing.T) {
	var total Money
	for range 10 {
		total += Cents(10)
	}
	if want := Euros(1); total != want {
		t.Errorf("ten times 0.10 = %s, want %s", total, want)
	}
}

func TestMoneyMarshalJSON(t *testing.T) {
	got, err := json.Marshal(struct {
		Amount Money `json:"amount"`
	}{Cents(1805)})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if want := `{"amount":"18.05"}`; string(got) != want {
		t.Errorf("json.Marshal() = %s, want %s", got, want)
	}
}
