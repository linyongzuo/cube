package pkg

import (
	"bufio"
	"bytes"
	"cube/gologger"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

func Contains(str string, slice []string) bool {
	//str是否在slice列表里面
	for _, value := range slice {
		if str == value {
			return true
		}
	}
	return false
}

func FileReader(filename string, trimSpace bool) []string {
	file, err := os.Open(filename)
	if err != nil {
		gologger.Errorf("Open file %s error, %v\n", filename, err)
	}
	defer file.Close()
	var content []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := ""
		if trimSpace {
			text = strings.TrimSpace(scanner.Text())
		} else {
			text = scanner.Text()
		}
		if text != "" {
			content = append(content, scanner.Text())
		}
	}
	return content
}

func ReadBytes(conn net.Conn) (result []byte, err error) {
	buf := make([]byte, 4096)
	for {
		count, err := conn.Read(buf)
		if err != nil {
			break
		}
		result = append(result, buf[0:count]...)
		if count < 4096 {
			break
		}
	}
	return result, err
}

func TrimName(name string) string {
	return strings.TrimSpace(strings.Replace(name, "\x00", "", -1))
}
func Bytes2Uint(bs []byte, endian byte) uint64 {
	var u uint64
	if endian == '>' {
		for i := 0; i < len(bs); i++ {
			u += uint64(bs[i]) << (8 * (len(bs) - i - 1))
		}
	} else {
		for i := 0; i < len(bs); i++ {
			u += uint64(bs[len(bs)-i-1]) << (8 * (len(bs) - i - 1))
		}
	}
	return u
}

func RemoveDuplicate(old []string) []string {
	result := make([]string, 0, len(old))
	temp := map[string]struct{}{}
	for _, item := range old {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func Bytes2StringUTF16(bs []byte) string {
	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	(*ptr).Len = ptr.Len / 2

	s := (*[]uint16)(unsafe.Pointer(&bs))
	return string(utf16.Decode(*s))
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func IsUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func StrXor(message string, keywords string) string {
	messageLen := len(message)
	keywordsLen := len(keywords)

	result := ""

	for i := 0; i < messageLen; i++ {
		result += string(message[i] ^ keywords[i%keywordsLen])
	}
	return result
}

func ValidIp(ip string) bool {
	addr := strings.Trim(ip, " ")
	regStr := `^(([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	if match, _ := regexp.MatchString(regStr, addr); match {
		return true
	}
	return false
}

func Split(r rune) bool {
	return strings.ContainsRune("://:", r)
}

func SameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}

func RemoveRepByMap(slc []string) []string {
	var result []string
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

func Subset(first, second []string) bool {
	set := make(map[string]int)
	for _, value := range second {
		set[value] += 1
	}

	for _, value := range first {
		if count, found := set[value]; !found {
			return false
		} else if count < 1 {
			return false
		} else {
			set[value] = count - 1
		}
	}
	return true
}

var mu *sync.RWMutex

func WriteToFile(f *os.File, output string) error {
	mu.Lock()
	_, err := f.WriteString(fmt.Sprintf("%s\n", output))
	mu.Unlock()
	if err != nil {
		return fmt.Errorf("[!] Unable to write to file %w", err)
	}
	return nil
}

func ByteToString(buf []byte) (string, error) {
	if utf8.Valid(buf) {
		return TrimName(string(buf)), nil
	} else {
		s1, _ := GbkToUtf8(buf)
		return TrimName(string(string(s1))), nil
	}
}
