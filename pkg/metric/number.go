package metric

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	regxNumOnly = regexp.MustCompile(`^[0-9]*$`)
)

type Result map[string]Number

// A Number represents a JSON number literal.
type Number string

func NewInt(n int) Number {
	return Number(fmt.Sprintf("%d", n))
}

func NewInt64(n int64) Number {
	return Number(fmt.Sprintf("%d", n))
}

func NewFloat32(n float32) Number {
	return Number(fmt.Sprintf("%f", n))
}

func NewFloat64(n float64) Number {
	return Number(fmt.Sprintf("%f", n))
}

func (n Number) MarshalJSON() ([]byte, error) {

	var val interface{} = strings.Clone(string(n))

	// JSON not support actual type, we detect if it contain dot then try to parse as float
	// if not contain, then it may integer if all is number.
	switch {
	case strings.Contains(string(n), "."):
		if f, err := strconv.ParseFloat(string(n), 64); err == nil {
			val = f
		}

	case regxNumOnly.MatchString(string(n)):
		if i, err := strconv.ParseInt(string(n), 10, 64); err == nil {
			val = i
		}

	default:
		if b, _err := strconv.ParseBool(string(n)); _err == nil {
			val = b
		}

	}

	return json.Marshal(val)
}
