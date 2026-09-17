package shipping

import (
	"fmt"
	"strconv"
)

type Money int64

func Euros(e int64) Money { return Money(e) * 100 }

func Cents(c int64) Money { return Money(c) }

func (m Money) String() string {
	v := int64(m)
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s%d.%02d", sign, v/100, v%100)
}

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(m.String())), nil
}
