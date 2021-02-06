package toolkit

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func RandomDateTimeString(strlen int, stype string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	var chars string
	if stype == "int" {
		chars = "0123456789"
	} else {
		chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	}
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	now := time.Now()
	str := now.Format("20060102150405")
	millis := now.UnixNano()
	millisStr := strconv.FormatInt(millis/1e6-millis/1e9*1000, 10)
	str = str + millisStr + string(result)
	return strings.ToUpper(str)
}

// GenerateRangeNum 生成一个区间范围的随机数
func GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max - min)
	randNum = randNum + min
	return randNum
}
