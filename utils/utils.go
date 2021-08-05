// Copyright 2021 CloudJ Company Limited. All rights reserved.

package utils

import (
	"archive/zip"
	"bytes"
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/utils/logs"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"math"
	"math/big"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

const letterAndDigit = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	n, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if n == nil {
		n = big.NewInt(time.Now().UnixNano())
	}
	rand.Seed(n.Int64())
}

func RandomStr(n int) string {
	r := make([]byte, n)
	for i := 0; i < n; i++ {
		r[i] = letterAndDigit[rand.Intn(len(letterAndDigit))]
	}
	return string(r)
}

func MaxUInt64(a uint64, b uint64) uint64 {
	if a >= b {
		return a
	} else {
		return b
	}
}

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func CheckPassword(password, hashedPassword string) (bool, error) {
	if hashedPassword == "" || password == "" {
		return false, nil
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func GlobMatch(pattern, str string) (bool, error) {
	// 该函数实现中，通配符不能匹配 "/"，因为这是路径分隔符
	return filepath.Match(pattern, str)
}

func LogLevel(verboseNum int) string {
	switch verboseNum {
	case 0:
		return "info"
	case 1:
		return "debug"
	default:
		return "trace"
	}
}

func RemoveDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func Md5String(ss ...string) string {
	hm := md5.New()
	for i := range ss {
		hm.Write([]byte(ss[i]))
	}
	return fmt.Sprintf("%x", hm.Sum(nil))
}

func Md5File(src io.ReadSeeker) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	checkSum := fmt.Sprintf("%x", hash.Sum(nil))
	return checkSum, nil
}

func GenProcKey(cwd string, cmdline string) string {
	return fmt.Sprintf("%s", Md5String(cwd, cmdline)[:16])
}

func SortedStringKV(m map[string]string) string {
	var (
		ks     = make([]string, 0, len(m))
		sorted = make([]string, 0, len(m))
	)
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		sorted = append(sorted, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(sorted, ",")
}

func FileExist(p string) bool {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func JoinUint(ids []uint, sep string) string {
	idsStr := make([]string, 0, len(ids))
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	return strings.Join(idsStr, sep)
}

func InArrayUint(arr []uint, v uint) bool {
	for i := range arr {
		if arr[i] == v {
			return true
		}
	}
	return false
}

func InArrayStr(arr []string, v string) bool {
	for i := range arr {
		if arr[i] == v {
			return true
		}
	}
	return false
}

func StrInArray(v string, arr ...string) bool {
	return InArrayStr(arr, v)
}

func UnzipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), f.Mode()); err != nil {
				return err
			}

			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func CheckRespCode(respCode int, code int) bool {
	return strings.HasSuffix(fmt.Sprintf("%d", respCode), fmt.Sprintf("%d", code))
}

func aesKey() []byte {
	sk := configs.Get().SecretKey
	if sk == "" {
		// "" 不是一个合法的 aes key，这里直接返回，等调用 aes.NewCipher() 时报错
		return []byte(sk)
	}
	if len(sk) == 32 {
		return []byte(sk)
	}
	return []byte(Md5String(sk))
}

func AesEncrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(aesKey())
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		return "", err
	}
	cipher.NewCFBEncrypter(block, iv).XORKeyStream(ciphertext[aes.BlockSize:],
		[]byte(plaintext))
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func AesDecrypt(d string) (string, error) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(d)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey())
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("cipher text too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	cipher.NewCFBDecrypter(block, iv).XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}

func MustJSON(v interface{}) []byte {
	bs, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bs
}

func MustJSONIndent(v interface{}, indent string) []byte {
	bs, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		panic(err)
	}
	return bs
}

func RecoverPanic(logger logs.Logger, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("panic: %v", r)
			logger.Errorf("%s", string(debug.Stack()))
		}
	}()

	fn()
}

func GenGuid(v string) string {
	guid := xid.New()
	guidStr := guid.String()
	if v != "" {
		guidStr = fmt.Sprintf("%s-%s", v, guidStr)
	}
	return guidStr
}

const (
	NUmStr  = "0123456789"
	CharStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	SpecStr = "+=-@#~,.[]()!%^*$"
)

func GenPasswd(length int, charset string) string {
	//初始化密码切片
	var passwd []byte = make([]byte, length, length)
	//源字符串
	var sourceStr string
	//判断字符类型,如果是数字
	if charset == "num" {
		sourceStr = NUmStr
		//如果选的是字符
	} else if charset == "char" {
		sourceStr = charset
		//如果选的是混合模式
	} else if charset == "mix" {
		sourceStr = fmt.Sprintf("%s%s", NUmStr, CharStr)
		//如果选的是高级模式
	} else if charset == "advance" {
		sourceStr = fmt.Sprintf("%s%s%s", NUmStr, CharStr, SpecStr)
	} else {
		sourceStr = NUmStr
	}

	//遍历，生成一个随机index索引,
	for i := 0; i < length; i++ {
		index := rand.Intn(len(sourceStr))
		passwd[i] = sourceStr[index]
	}
	return string(passwd)
}

func UintIsContain(items []uint, item uint) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

// RetryFunc 通用重试函数，
// param max: 最大重试次数，传 0 表示一直重试
// param maxDelay: 最大重试等待时长
// param run: 重试执行的函数，入数 retryN 为当前重试次数(0 base)，返回值分别为(继续重试?, error)
// return: 返回值为 run() 函数返回的 error
func RetryFunc(max int, maxDelay time.Duration, run func(retryN int) (retry bool, err error)) error {
	retryCount := 0
	maxRetry := max // 最大重试次数(不含第一次)
	for {
		if retry, err := run(retryCount); retry {
			retryCount += 1
			if maxRetry > 0 && retryCount > maxRetry {
				return err
			}

			delay := time.Duration(retryCount) * 2 * time.Second
			if delay > maxDelay { // 最大重试等待时长 10s
				delay = maxDelay
			}
			time.Sleep(delay)
			continue
		} else {
			return err
		}
	}
}

func TaskLogMessage(format string, args ...interface{}) string {
	return fmt.Sprintf(consts.IacTaskLogPrefix+format, args...)
}

func TaskLogMsgBytes(format string, args ...interface{}) []byte {
	return []byte(TaskLogMessage(format, args...))
}

// LimitOffset2Page
// offset 必须为 limit 的整数倍，否则会 panic
// page 从 1 开始
func LimitOffset2Page(limit int, offset int) (page int) {
	if limit <= 0 {
		return 1
	}

	if offset%limit != 0 {
		panic(fmt.Errorf("LimitOffset2Page: offset(%d) %% limit(%d) != 0", offset, limit))
	}
	return (offset / limit) + 1
}

// PageSize2Offset page 从 1 开始
func PageSize2Offset(page int, pageSize int) (offset int) {
	if page <= 1 {
		return 0
	}
	return (page - 1) * pageSize
}

// GenQueryURL url拼接
func GenQueryURL(address string, path string, params url.Values) string {
	address = GetUrl(address)
	if params != nil {
		return fmt.Sprintf("%s%s?%s", address, path, params.Encode())
	} else {
		return fmt.Sprintf("%s%s", address, path)
	}
}

func ShortContainerId(id string) string {
	if len(id) < 12 {
		return id
	}
	return id[:12]
}

// GetBoolEnv 判断环境变量 bool 值
func GetBoolEnv(key string, _default bool) bool {
	val := os.Getenv(key)
	if StrInArray(val, "off", "false", "0") {
		// 明确设置了 "false" 值则返回 false
		return false
	} else if StrInArray(val, "on", "true", "1") {
		// 明确设置了 "true" 值则返回 true
		return true
	}
	// 其他情况返回默认值
	return _default
}

func JoinURL(address string, path ...string) string {
	elems := append([]string{address}, path...)
	for i := range elems {
		elems[i] = strings.Trim(elems[i], "/")
	}
	return strings.Join(elems, "/")
}

// SprintTemplate 用模板参数格式化字符串
func SprintTemplate(format string, data interface{}) (str string) {
	if tmpl, err := template.New("").Parse(format); err != nil {
		return format
	} else {
		var msg bytes.Buffer
		tmpl.Execute(&msg, data)
		return msg.String()
	}
}

func SliceEqualStr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func GetUUID() (string, error) {
	u2, err := uuid.NewV4()
	if err != nil {
		logs.Get().Errorf("Something went wrong: %s", err)
		return "", err
	}
	return u2.String(), nil
}

// FirstValueStr 获取参数列表中第一个非空的字符串
func FirstValueStr(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func SetGinMode() {
	if mode := os.Getenv(gin.EnvGinMode); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}
