package idconvertor

import (
	"math"

	"github.com/illacloud/illa-builder-backend/src/utils/config"
)

var table = ""
var tr = map[string]int{}
var s = []int{11, 10, 3, 8, 4, 6}
var xor = 111111111
var add = 9999999999

func init() {
	conf := config.GetInstance()
	table = conf.GetIDTable()
	tableByte := []byte(table)
	for i := 0; i < 58; i++ {
		tr[string(tableByte[i])] = i
	}
}

func ConvertStringToInt(bv string) int {
	var r int
	arr := []rune(bv)

	for i := 0; i < 6; i++ {
		r += tr[string(arr[s[i]])] * int(math.Pow(float64(58), float64(i)))
	}
	return (r - add) ^ xor
}

func ConvertIntToString(av int) string {
	x := (av ^ xor) + add
	r := []string{"I", "L", "A", " ", " ", "4", " ", "1", " ", "7", " ", " "}
	for i := 0; i < 6; i++ {
		r[s[i]] = string(table[int(math.Floor(float64(x/int(math.Pow(float64(58), float64(i))))))%58])
	}
	var result string
	for i := 0; i < 12; i++ {
		result += r[i]
	}
	return result
}
