package strutil

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mozillazg/go-pinyin"
	"github.com/pborman/uuid"
	"github.com/wednesdaysunny/onerpc/eco/inter/common"
	"github.com/wednesdaysunny/onerpc/eco/inter/toolkit/intutil"
	"github.com/wednesdaysunny/onerpc/eco/inter/toolkit/sliceutil"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	UriRegexRule = `(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`
)

var (
	uriRep   = regexp.MustCompile(UriRegexRule)
	pinyinOc = pinyin.NewArgs()
)

func init() {
	pinyinOc.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}
}

func GenContent(content string, length int) string {
	result := []rune(content)
	if len(result) > length {
		result = result[:length]
	}
	return strings.TrimSpace(string(result))
}

var charArr = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
	"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	"u", "v", "w", "x", "y", "z", "A", "B", "C", "D",
	"E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
	"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
	"Y", "Z"}

var base64Images = []string{
	"data:image/jpeg;base64",
	"data:image/png;base64",
	"data:image/gif;base64",
	"data:text/plain;base64",
}

func FromInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func FromUInt64(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func FromFloat64(i float64) string {
	return strconv.FormatFloat(i, 'f', 2, 64)
}

func ToInt64(str string) int64 {
	if b, err := strconv.ParseInt(str, 10, 64); err != nil {
		return 0
	} else {
		return b
	}
}

func ToInt32(str string) int32 {
	/*
		int means at least 32 ,not an alias for int32
	*/
	if b, err := strconv.ParseInt(str, 10, 32); err == nil {
		return int32(b)
	}
	return 0
}

/*
 * 字符串中ID以,分割的ids
 */
func ToInt64Arr(str string) []int64 {

	res := make([]int64, 0)
	idSplit := strings.Split(str, common.CommaSeparator)

	for _, val := range idSplit {
		valint64 := intutil.ToInt64(val, 0)
		if valint64 > 0 {
			res = append(res, valint64)
		}
	}
	return res
}

// 将string数组转成int64 数组,转换失败的忽略

func StrArrToInt64Arr(strs []string) []int64 {
	res := make([]int64, 0)
	for _, str := range strs {
		r, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			res = append(res, r)
		}
	}
	return res
}

// 将int64数组转成 string数组,转换失败的忽略
func Int64ArrToStrArr(ints []int64) []string {
	res := make([]string, 0)
	for _, v := range ints {
		res = append(res, strconv.FormatInt(v, 10))
	}
	return res
}

// 将int数组转成 string数组,转换失败的忽略
func IntArrToStrArr(ints []int) []string {
	res := make([]string, 0)
	for _, v := range ints {
		res = append(res, strconv.Itoa(v))
	}
	return res
}

/*
 * 字符串中ID以,分割的ids
 */
func ToStringArr(str string) []string {
	res := make([]string, 0)
	idSplit := strings.Split(str, common.CommaSeparator)

	for _, val := range idSplit {
		if len(val) > 0 {
			res = append(res, strings.TrimSpace(val))
		}
	}
	return res
}

func AddSqlLikePreAndSubfix(in string) string {
	return "%" + in + "%"
}

func AddSqlLikePrefix(in string) string {
	return "%" + in
}

func AddSqlLikeSubfix(in string) string {
	return in + "%"
}

func ArrToString(arr []string) string {
	return strings.Join(arr, common.CommaSeparator)
}

func FromInt(val int) string {
	return strconv.Itoa(val)
}

func ToInt(str string) int {
	if b, err := strconv.ParseInt(str, 10, 64); err != nil {
		return 0
	} else {
		return int(b)
	}
}

func FromByte(val byte) string {
	return strconv.Itoa(int(val))
}

func ToByte(str string) byte {
	if b, err := strconv.ParseInt(str, 10, 8); err != nil {
		return 0
	} else {
		return byte(b)
	}
}

func FromBool(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func ToBool(s string) bool {
	lower := strings.ToLower(s)
	if lower == "true" || lower == "yes" || lower == "1" {
		return true
	}
	return false
}

func ToStrList(str string) []string {
	if len(str) <= 1 {
		return nil
	} else if str == "null" {
		return nil
	}
	var v []string
	json.Unmarshal([]byte(str), &v)
	return v
}

func ToObject(obj interface{}, str string) {
	json.Unmarshal([]byte(str), obj)
}

func UnmarshalToInt64Array(str string) (res []int64) {
	ToObject(&res, str)
	return
}

func FromStrList(strs []string) string {
	if strs == nil {
		return "[]"
	}
	data, _ := json.Marshal(&strs)
	return string(data)
}

func FromObject(obj interface{}) string {
	data, _ := json.Marshal(obj)
	return string(data)
}

func FromObjectPretty(obj interface{}) string {
	data, _ := json.MarshalIndent(obj, "", "\t")
	return string(data)
}

func ToSqlNullString(key string, isValid bool) sql.NullString {
	res := sql.NullString{
		String: key,
		Valid:  isValid,
	}
	return res
}

func RandomString(length int) string {
	if length < 1 {
		return ""
	}
	result := make([]string, length)
	for i := 0; i < length; i++ {
		result = append(result, charArr[rand.Intn(61)])
	}
	return strings.Join(result, "")
}

func Sub(str string, begin, length int) string {
	rs := []rune(str)
	lth := len(rs)

	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[begin:end])
}

func ToLike(str string) string {
	return fmt.Sprintf("%%%s%%", str)
}

func UUID() string {
	return uuid.New()
}

func Diff(arr1 []string, arr2 []string) []string {
	res := make([]string, 0)
	for _, val1 := range arr1 {
		exist := false
		for _, val2 := range arr2 {
			if val1 == val2 {
				exist = true
				break
			}
		}
		if !exist {
			res = append(res, val1)
		}
	}
	return res
}

func FromChinese(str string) string {
	res := strings.Join(pinyin.LazyPinyin(str, pinyinOc), "")
	if IsEmpty(res) {
		res = str
	}
	return res
}

func IsEmpty(str string) bool {
	return str == ""
}

func AsciiPinyin(s string) string {
	options := pinyin.NewArgs()
	options.Fallback = func(r rune, options pinyin.Args) []string {
		if r < 128 {
			return []string{string(r)}
		} else {
			return []string{}
		}
	}
	result := strings.Join(pinyin.LazyPinyin(s, options), "")
	return strings.ToLower(result)
}

func AsciiPinyinLower(s string) string {
	s = strings.ToLower(AsciiPinyin(s))
	r := regexp.MustCompile(`[^a-z0-9]`)
	s = r.ReplaceAllString(s, "")
	return s
}

func ProcessMobile(mobile string) string {
	if len(mobile) >= 1 && mobile[0:1] != "+" {
		return mobile
	}
	if len(mobile) >= 3 && mobile[0:3] == "+86" {
		return mobile[3:]
	}
	return mobile
}

func TrimSpace(str string) string {
	return strings.Trim(str, " ")
}

func StrInArray(str string, strs []string) bool {
	ina := false
	for _, st := range strs {
		if st == str {
			ina = true
			break
		}
	}
	return ina
}

func Int64InArray(int int64, ints []int64) bool {
	ina := false
	for _, st := range ints {
		if st == int {
			ina = true
			break
		}
	}
	return ina
}

func Md5(str string) string {
	h := md5.New()
	return hex.EncodeToString(h.Sum([]byte(str)))
}

func Get32MD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func Md5Object(obj interface{}) string {
	data, _ := json.Marshal(obj)
	return Md5(string(data))
}

func DecodePercent(s string) string {
	if decoded, err := url.PathUnescape(s); err != nil {
		return ""
	} else {
		return decoded
	}
}

func DecodeCsvArg(s string) []string {
	vs := strings.Split(DecodePercent(s), string(44))
	vs = sliceutil.RemoveEmptyStrings(vs)
	vs = sliceutil.RemoveDuplicateStrings(vs)
	sort.Strings(vs)
	return vs
}

func HMacMd5Sign(key string, str string) string {
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func ContentHasBase64Image(con string) bool {
	for _, val := range base64Images {
		if strings.Index(con, val) > -1 {
			return true
		}
	}
	return false
}

func GbkLen(s string) int {
	data, err := ioutil.ReadAll(
		transform.NewReader(bytes.NewReader([]byte(s)),
			simplifiedchinese.GB18030.NewEncoder()))
	if err != nil {
		return len(s)
	} else {
		return len(data)
	}
}

func CompileUris(str string) []string {
	res := uriRep.FindAllString(str, -1)

	return res
}

func ToFloat64(str string) float64 {
	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return i
}

func ComplementIdPrefix(str, fix string) int64 {
	if len(str) >= len(fix) {
		return ToInt64(str)
	}
	str = str + strings.Repeat("0", len(fix)-len(str))

	return ToInt64(str)
}

func GenEncodedKeyFromObject(r interface{}) string {

	return base64.RawURLEncoding.EncodeToString([]byte(FromObject(r)))
}

func HighLight(content, query string) string {
	var (
		rst        string
		index      int
		queryLen   = len(query)
		tmpContent = content
	)

	index = strings.Index(tmpContent, query)
	for index >= 0 {
		for i, v := range tmpContent {
			if i >= index {
				break
			}
			rst += string(v)
		}
		rst = rst + "<em>" + query + "</em>"
		tmpContent = tmpContent[index+queryLen:]
		index = strings.Index(tmpContent, query)
	}
	rst += tmpContent
	return rst
}

func HighLightEn(content, query string, toLower bool) string {
	var (
		rst        string
		index      int
		queryLen   = len(query)
		tmpContent = content
	)

	index = strings.Index(ToLower(tmpContent, toLower), ToLower(query, toLower))
	for index >= 0 {
		for i, v := range tmpContent {
			if i >= index {
				break
			}
			rst += string(v)
		}
		rst = rst + "<em>" + tmpContent[index:index+queryLen] + "</em>"
		tmpContent = tmpContent[index+queryLen:]
		index = strings.Index(ToLower(tmpContent, toLower), ToLower(query, toLower))
	}
	rst += tmpContent
	return rst
}

func ToLower(content string, toLower bool) string {
	if toLower {
		return strings.ToLower(content)
	}
	return content
}

func Int64ArrToString(arr []int64) string {
	byteArrString, _ := json.Marshal(&arr)

	return string(byteArrString[:])

}

func StrToValueType(s string, valueType string) interface{} {
	var value interface{}
	if valueType == "float64" || valueType == "float" {
		value = ToFloat64(s)
	} else if valueType == "numeric" {
		value = ToInt64(s)
	} else if valueType == "boolean" {
		value = ToBool(s)
	} else {
		value = s
	}
	return value
}

func Split(s, sep string) []string {
	if IsEmpty(strings.TrimSpace(s)) {
		return nil
	}
	return strings.Split(s, sep)
}

type QiniuImageInfo struct {
	Uri    string `json:"uri"`
	Size   int64  `json:"size"`
	Format string `json:"format"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

func GetQiniuImageInfo(u string) *QiniuImageInfo {
	var res = &QiniuImageInfo{
		Uri: u,
	}
	resp, err := http.Get(u + "?imageInfo")
	if err != nil {
		return res

	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(s, res)
	if err != nil {
		return res
	}
	res.Uri = u
	return res
}

func IsQiniuDocExist(u string) bool {
	var res = &struct {
		Error string `json:"error"`
	}{}
	resp, err := http.Get(u)
	if err != nil {
		return true

	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(s, res)
	if err != nil {
		return true
	}
	if res.Error == "Document not found" {
		return false
	}
	return true
}

var defaultQiniuHostMap = map[string]string{
	"p5j1itukc.bkt.clouddn.com": "alliance-img.xuangubao.cn",
}

func ReplaceQiniuDefaultHost(u string) string {
	for k, v := range defaultQiniuHostMap {
		u = strings.Replace(u, k, v, -1)
	}
	return u
}
