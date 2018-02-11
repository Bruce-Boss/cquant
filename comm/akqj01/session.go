package akqj01

import (
	"time"
	"math/rand"
	"github.com/datochan/gcom/bytes"
	"github.com/datochan/gcom/logger"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
)


type C10JQKASession struct{
	R map[string] int
	A string
	Data  []int
	basicFields []byte
}

func New10JQKASession() *C10JQKASession {
	r := map[string]int{}
	a := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for x := 0; x < 64; x++ {
		r[string(a[x])] = x
	}

	return &C10JQKASession{ R: r, A: a,
		Data:[]int{1450088886/*随机数*/, 1518154760/*服务器时间*/, 1518265962 /*当前时间*/,
			518136465 /*navigator.userAgent strhash*/, 7, 10, 4, 0, 0, 0, 0, 0, 0, 3756, 0, 0, 2, 2},
		basicFields:[]byte{4, 4, 4, 4, 1, 1, 1, 3, 2, 2, 2, 2, 2, 2, 2, 4, 2, 1}}
}

func (impl *C10JQKASession) CheckSum(n []byte) byte{
	var t byte
	var e int
	for len(n) > e {
		t = (t << 5) - t + n[e]
		e++
	}

	return t & 255
}

func (impl *C10JQKASession) random() float32 {
	return rand.Float32() * 4294967295.0
}

/**
function S(data, dataIdx, out, ourIdx, cs) {
	var count = data["length"];
	for (; count > dataIdx;){
		out[ourIdx++] = data[dataIdx++] ^ cs & 255;
		cs = ~ (cs * 131);
	}
}
 */
func (impl *C10JQKASession) S(data []byte, dataIdx int, cs byte) []byte {
	var out []byte

	for count := len(data); count > dataIdx; dataIdx ++{
		out = append(out, data[dataIdx] ^ cs & 255)
		cs = ^(cs * 131)
	}

	return out
}

func (impl *C10JQKASession) base64Encode(o []byte) string {
	var resultList []byte

	for idx := 0;len(o) > idx; {
		m1 := int(o[idx]) << 16
		idx++

		m2 := 0
		if len(o) > idx { m2 = int(o[idx]) << 8; idx++}

		m3 := 0
		if len(o) > idx { m3 = int(o[idx]); idx++}

		var m = m1 | m2 | m3

		resultList = append(resultList, impl.A[m >> 18], impl.A[m >> 12 & 63], impl.A[m >> 6 & 63], impl.A[m & 63])
	}

	return bytes.BytesToString(resultList)
}

func (impl *C10JQKASession) base64Decode(o string) []byte {
	var l []byte
	idx := 0

	for len(o) > idx {
		p1 := impl.R[string(o[idx])]
		p18 := p1 << 18
		idx++

		p12 := 0
		if len(o) > idx { p2 := impl.R[string(o[idx])]; p12 = p2 << 12; idx++ }

		p6 := 0
		if len(o) > idx { p3 := impl.R[string(o[idx])]; p6 = p3 << 6; idx++ }

		p4 := 0
		if len(o) > idx { p4 = impl.R[string(o[idx])]; idx++ }

		p := p18 | p12 | p6 | p4

		l = append(l, byte(p >> 16), byte(p >> 8 & 255), byte(p & 255))
	}

	return l
}

// todo: 未完成, 用不到就没有再翻译
func (impl *C10JQKASession) Decode(r string) []int {
	a := impl.base64Decode(r)

	//if a[0] != 2 {
	//	fmt.Println("error...")
	//	return nil
	//}

	T := a[1]
	b := impl.S(a, 2, T)
	sum := impl.CheckSum(b)

	if sum == T {
		return impl.DecodeBuffer(b)
	} else {
		return nil
	}
}

func (impl *C10JQKASession) Encode() string {
	data := impl.ToBuffer()
	t := impl.CheckSum(data)
	e := []byte{2, t}
	e = bytes.BytesCombine(e, impl.S(data, 0, t))
	return impl.base64Encode(e)
}

func (impl *C10JQKASession) ToBuffer() []byte {
	var resultList []byte

	fieldCount := len(impl.basicFields)

	for idx := 0; idx < fieldCount; idx++ {
		itemData := impl.Data[idx]
		itemField := impl.basicFields[idx]
		tmpBuffer := make([]byte, itemField)

		for fieldIdx := itemField - 1; itemField > 0; fieldIdx-- {
			tmpBuffer[fieldIdx] = byte(itemData & 255)
			itemData >>= 8
			itemField--
		}

		resultList = bytes.BytesCombine(resultList, tmpBuffer)
	}

	return resultList
}

func (impl *C10JQKASession) DecodeBuffer(dataList []byte) []int {
	idx := 0
	fieldCount := len(impl.basicFields)

	for i := 0; fieldCount > i; i++ {
		dataValue := 0

		itemField := impl.basicFields[i]

		for {
			dataValue = (dataValue << 8) + int(dataList[idx])
			idx++
			if itemField--; itemField <= 0 {
				break
			}
		}

		impl.Data[i] = dataValue
	}

	return impl.Data
}

func (impl *C10JQKASession) UpdateServerTime() {
	t := time.Now()
	curTimestamp := int(t.Unix())

	resp, err := http.Get(fmt.Sprintf("https://s.thsi.cn/js/chameleon/time.%d.js", curTimestamp / 1200))
	if err != nil {
		// handle error
		logger.Fatal(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	jsCnt := string(body)
	// var TOKEN_SERVER_TIME=1518313168.121;// server time
	start := strings.Index(string(body), "=")+1
	end := strings.Index(string(body), ";")
	st := jsCnt[start:end]
	serverTime, _ := strconv.ParseFloat(st, 64)

	impl.Data[0] = int(impl.random())
	impl.Data[1] = int(serverTime)
	impl.Data[2] = int(curTimestamp)
}
