package values

import (
	"strconv"

	"github.com/tsavola/wag/internal/types"
)

func I32(x interface{}) uint32 {
	s := x.(string)

	signed64, err := strconv.ParseInt(s, 0, 32)
	if err == nil {
		return uint32(signed64)
	}

	unsigned64, err := strconv.ParseUint(s, 0, 32)
	if err == nil {
		return uint32(unsigned64)
	}

	panic(err)
}

func I64(x interface{}) uint64 {
	s := x.(string)

	signed64, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		return uint64(signed64)
	}

	unsigned64, err := strconv.ParseUint(s, 0, 64)
	if err == nil {
		return unsigned64
	}

	panic(err)
}

func F32(x interface{}) float32 {
	s := x.(string)

	value64, err := strconv.ParseFloat(s, 32)
	if err == nil {
		return float32(value64)
	}

	panic(err)
}

func F64(x interface{}) float64 {
	s := x.(string)

	value64, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return value64
	}

	panic(err)
}

func Parse(t types.T, x interface{}) interface{} {
	switch t {
	case types.I32:
		return I32(x)

	case types.I64:
		return I64(x)

	case types.F32:
		return F32(x)

	case types.F64:
		return F64(x)

	default:
		panic(t)
	}
}