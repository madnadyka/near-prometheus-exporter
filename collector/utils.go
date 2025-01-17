package collector

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
)

func GetStakeFromString(s string) float64 {
	if len(s) <= 19 {
		return 0
	}
	l := len(s) - 19
	v, err := strconv.ParseFloat(s[0:l], 64)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return v / math.Pow(10, 5)
}

func HashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
