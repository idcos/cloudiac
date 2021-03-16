package utils

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"cloudiac/utils/logs"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const letterAndDigit = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	commonKey = []byte("monitorSecretKey")
)

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

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

func FileExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
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

func AesEncrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(commonKey)
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
	return hex.EncodeToString(ciphertext), nil
}

func AesDecrypt(d string) string {
	ciphertext, err := hex.DecodeString(d)
	logger := logrus.WithField("func", "AesDecrypt")
	if err != nil {
		logger.Errorln(err)
		return ""
	}
	block, err := aes.NewCipher(commonKey)
	if err != nil {
		logger.Errorln(err)
		return ""
	}
	if len(ciphertext) < aes.BlockSize {
		logger.Errorln(errors.New("cipher text too short"))
		return ""
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	cipher.NewCFBDecrypter(block, iv).XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext)
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
