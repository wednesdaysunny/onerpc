package intutil

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestToInt32(t *testing.T) {
	value := ToInt32("1", 2)
	if value != 1 {
		t.Error("等于1撒")
	}
	value = ToInt32("sdf", 2)
	if value != 2 {
		t.Error("等于2撒")
	}

}

func TestFloatToPriceInt64(t *testing.T) {
	fmt.Println(FloatToPriceInt64(21.176))
}

func TestJiecheng(t *testing.T) {
	var n int64 = 100
	var c int64 = 1
	var i int64 = 1
	for ; i <= n; i++ {
		fmt.Println(c, "  ", i)
		c = c * i
	}
	fmt.Println(c)
}

func TestAA(t *testing.T) {
	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 20; i++ {
		fmt.Println(r1.Intn(20))
	}

}

func TestIndexInt64(t *testing.T) {
	limit := 2
	a := []int64{6, 3, 4, 8, 2, 1}
	index := IndexInt64(a, 4)
	fmt.Println()
	if (index + limit + 2) > len(a) {
		fmt.Println(a[index+1:])
	} else {
		fmt.Println(a[(index + 1):(index + limit + 1)])
	}
	a1 := time.Now().UnixNano()
	a2 := time.Now().Unix()
	fmt.Println(a1 / 1000000000)
	fmt.Println(a2)
}

func TestRandom(t *testing.T) {
	a1 := time.Now().UnixNano()
	a2 := a1 + 404476284
	fmt.Println(a1)
	fmt.Println(a2)
}
