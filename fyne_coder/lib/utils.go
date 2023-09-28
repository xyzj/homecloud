package data

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	// OSNAME from runtime
	OSNAME = runtime.GOOS
	// OSARCH from runtime
	OSARCH = runtime.GOARCH
	// DateTimeFormat yyyy-mm-dd hh:MM:ss
	DateTimeFormat = "2006-01-02 15:04:05"
	// DateOnlyFormat yyyy-mm-dd hh:MM:ss
	DateOnlyFormat = "2006-01-02"
	// TimeOnlyFormat yyyy-mm-dd hh:MM:ss
	TimeOnlyFormat = "15:04:05"
	// LongTimeFormat 含日期的日志内容时间戳格式 2006/01/02 15:04:05.000
	LongTimeFormat = "2006-01-02 15:04:05.000"
	// ShortTimeFormat 无日期的日志内容时间戳格式 15:04:05.000
	ShortTimeFormat = "15:04:05.000"
	// FileTimeFormat 日志文件命名格式 060102
	FileTimeFormat = "060102" // 日志文件命名格式
)
const (
	// CryptoMD5 md5算法
	CryptoMD5 = iota
	// CryptoSHA256 sha256算法
	CryptoSHA256
	// CryptoSHA512 sha512算法
	CryptoSHA512
	// CryptoHMACSHA1 hmacsha1摘要算法
	CryptoHMACSHA1
	// CryptoHMACSHA256 hmacsha256摘要算法
	CryptoHMACSHA256
	// CryptoAES128CBC aes128cbc算法
	CryptoAES128CBC
	// CryptoAES128CFB aes128cfb算法
	CryptoAES128CFB
	// CryptoAES192CBC aes192cbc算法
	CryptoAES192CBC
	// CryptoAES192CFB aes192cfb算法
	CryptoAES192CFB
	// CryptoAES256CBC aes256cbc算法
	CryptoAES256CBC
	// CryptoAES256CFB aes256cfb算法
	CryptoAES256CFB
)

// CryptoWorker 序列化或加密管理器
type CryptoWorker struct {
	cryptoType   byte
	cryptoHash   hash.Hash
	cryptoLocker sync.Mutex
	cryptoIV     []byte
	cryptoBlock  cipher.Block
}

// GetNewCryptoWorker 获取新的序列化或加密管理器
// md5,sha256,sha512初始化后直接调用hash
// hmacsha1初始化后需调用SetSignKey设置签名key后调用hash
// aes加密算法初始化后需调用SetKey设置key和iv后调用Encrypt，Decrypt
func GetNewCryptoWorker(cryptoType byte) *CryptoWorker {
	h := &CryptoWorker{
		cryptoType: cryptoType,
	}
	switch cryptoType {
	case CryptoMD5:
		h.cryptoHash = md5.New()
	case CryptoSHA256:
		h.cryptoHash = sha256.New()
	case CryptoSHA512:
		h.cryptoHash = sha512.New()
	case CryptoHMACSHA1:
		h.cryptoHash = hmac.New(sha1.New, []byte{})
	case CryptoHMACSHA256:
		h.cryptoHash = hmac.New(sha256.New, []byte{})
	}
	return h
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5Unpadding(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

// SetKey 设置aes-key,iv
func (h *CryptoWorker) SetKey(key, iv string) error {
	switch h.cryptoType {
	case CryptoHMACSHA1:
		h.cryptoHash = hmac.New(sha1.New, Bytes(key))
	case CryptoHMACSHA256:
		h.cryptoHash = hmac.New(sha256.New, Bytes(key))
	case CryptoAES128CBC:
		if len(key) < 16 || len(iv) < 16 {
			return fmt.Errorf("key length must be longer than 16, and the length of iv must be 16")
		}
		h.cryptoBlock, _ = aes.NewCipher(Bytes(key)[:16])
		h.cryptoIV = Bytes(iv)[:16]
	case CryptoAES192CBC:
		if len(key) < 24 || len(iv) < 16 {
			return fmt.Errorf("key length must be longer than 24, and the length of iv must be 16")
		}
		h.cryptoBlock, _ = aes.NewCipher(Bytes(key)[:24])
		h.cryptoIV = Bytes(iv)[:16]
	case CryptoAES256CBC:
		if len(key) < 32 || len(iv) < 16 {
			return fmt.Errorf("key length must be longer than 32, and the length of iv must be 16")
		}
		h.cryptoBlock, _ = aes.NewCipher(Bytes(key)[:32])
		h.cryptoIV = Bytes(iv)[:16]
	case CryptoAES128CFB:
		if len(key) < 16 || len(iv) < 16 {
			return fmt.Errorf("key length must be longer than 16, and the length of iv must be 16")
		}
		h.cryptoBlock, _ = aes.NewCipher(Bytes(key)[:16])
		h.cryptoIV = Bytes(iv)[:16]
	case CryptoAES192CFB:
		if len(key) < 24 || len(iv) < 16 {
			return fmt.Errorf("key length must be longer than 24, and the length of iv must be 16")
		}
		h.cryptoBlock, _ = aes.NewCipher(Bytes(key)[:24])
		h.cryptoIV = Bytes(iv)[:16]
	case CryptoAES256CFB:
		if len(key) < 32 || len(iv) < 16 {
			return fmt.Errorf("key length must be longer than 32, and the length of iv must be 16")
		}
		h.cryptoBlock, _ = aes.NewCipher(Bytes(key)[:32])
		h.cryptoIV = Bytes(iv)[:16]
	default:
		return fmt.Errorf("not yet supported")
	}
	return nil
}

// Encrypt 加密
func (h *CryptoWorker) Encrypt(s string) string {
	// h.cryptoLocker.Lock()
	// defer h.cryptoLocker.Unlock()
	if len(h.cryptoIV) == 0 {
		return ""
	}
	switch h.cryptoType {
	case CryptoAES128CBC, CryptoAES192CBC, CryptoAES256CBC:
		content := pkcs5Padding(Bytes(s), h.cryptoBlock.BlockSize())
		crypted := make([]byte, len(content))
		cipher.NewCBCEncrypter(h.cryptoBlock, h.cryptoIV).CryptBlocks(crypted, content)
		return base64.StdEncoding.EncodeToString(crypted)
	case CryptoAES128CFB, CryptoAES192CFB, CryptoAES256CFB:
		crypted := make([]byte, aes.BlockSize+len(s))
		cipher.NewCFBEncrypter(h.cryptoBlock, h.cryptoIV).XORKeyStream(crypted[aes.BlockSize:], Bytes(s))
		return base64.StdEncoding.EncodeToString(crypted)
	}
	return ""
}

// EncryptNoTail 加密，去掉base64尾巴的=符号
func (h *CryptoWorker) EncryptNoTail(s string) string {
	return strings.Replace(h.Encrypt(s), "=", "", -1)
}

// Decrypt 解密
func (h *CryptoWorker) Decrypt(s string) string {
	// h.cryptoLocker.Lock()
	// defer h.cryptoLocker.Unlock()
	defer func() { recover() }()
	if len(h.cryptoIV) == 0 {
		return ""
	}

	if x := 4 - len(s)%4; x != 4 {
		for i := 0; i < x; i++ {
			s += "="
		}
	}
	msg, _ := base64.StdEncoding.DecodeString(s)
	switch h.cryptoType {
	case CryptoAES128CBC, CryptoAES192CBC, CryptoAES256CBC:
		decrypted := make([]byte, len(msg))
		cipher.NewCBCDecrypter(h.cryptoBlock, h.cryptoIV).CryptBlocks(decrypted, msg)
		return String(pkcs5Unpadding(decrypted))
	case CryptoAES128CFB, CryptoAES192CFB, CryptoAES256CFB:
		msg = msg[aes.BlockSize:]
		cipher.NewCFBDecrypter(h.cryptoBlock, h.cryptoIV).XORKeyStream(msg, msg)
		return String(msg)
	}
	return ""
}

// Hash 计算序列
func (h *CryptoWorker) Hash(b []byte) string {
	h.cryptoLocker.Lock()
	defer h.cryptoLocker.Unlock()
	switch h.cryptoType {
	case CryptoMD5, CryptoSHA256, CryptoSHA512, CryptoHMACSHA1, CryptoHMACSHA256:
		h.cryptoHash.Reset()
		h.cryptoHash.Write(b)
		return fmt.Sprintf("%x", h.cryptoHash.Sum(nil))
	}
	return ""
}

// HashB64 返回base64编码格式
func (h *CryptoWorker) HashB64(b []byte) string {
	h.cryptoLocker.Lock()
	defer h.cryptoLocker.Unlock()
	switch h.cryptoType {
	case CryptoMD5, CryptoSHA256, CryptoSHA512, CryptoHMACSHA1, CryptoHMACSHA256:
		h.cryptoHash.Reset()
		h.cryptoHash.Write(b)
		return base64.StdEncoding.EncodeToString(h.cryptoHash.Sum(nil))
	}
	return ""
}

// GetMD5 生成32位md5字符串
func GetMD5(text string) string {
	ctx := md5.New()
	ctx.Write(Bytes(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// HashData 计算hash
func HashData(b []byte, cryptoType byte) string {
	switch cryptoType {
	case CryptoMD5:
		return fmt.Sprintf("%x", md5.Sum(b))
	case CryptoSHA256:
		return fmt.Sprintf("%x", sha256.Sum256(b))
	case CryptoSHA512:
		return fmt.Sprintf("%x", sha512.Sum512(b))
	}
	return ""
}

// ReverseString ReverseString
func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// CodeString 编码字符串
func CodeString(s string) string {
	x := byte(rand.Int31n(126) + 1)
	l := len(s)
	salt := GetRandomASCII(int64(l))
	var y, z bytes.Buffer
	for _, v := range Bytes(s) {
		y.WriteByte(v + x)
	}
	zz := y.Bytes()
	var c1, c2 int
	z.WriteByte(x)
	for i := 1; i < 2*l; i++ {
		if i%2 == 0 {
			z.WriteByte(salt[c1])
			c1++
		} else {
			z.WriteByte(zz[c2])
			c2++
		}
	}
	a := base64.StdEncoding.EncodeToString(z.Bytes())
	a = ReverseString(a)
	a = SwapCase(a)
	a = strings.Replace(a, "=", "", -1)
	return a
}

// DecodeString 解码混淆字符串，兼容python算法
func DecodeString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return ""
	}
	s = ReverseString(SwapCase(s))
	if x := 4 - len(s)%4; x != 4 {
		for i := 0; i < x; i++ {
			s += "="
		}
	}
	if y, ex := base64.StdEncoding.DecodeString(s); ex == nil {
		var ns bytes.Buffer
		x := y[0]
		for k, v := range y {
			if k%2 != 0 {
				ns.WriteByte(v - x)
			}
		}
		return ns.String()
	}
	return ""
}

// GetRandomASCII 获取随机ascII码字符串
func GetRandomASCII(l int64) []byte {
	var rs bytes.Buffer
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < l; i++ {
		rs.WriteByte(byte(rand.Int31n(255) + 1))
	}
	return rs.Bytes()
}

// SwapCase swap char case
func SwapCase(s string) string {
	var ns bytes.Buffer
	for _, v := range s {
		if v >= 65 && v <= 90 {
			ns.WriteString(string(v + 32))
		} else if v >= 97 && v <= 122 {
			ns.WriteString(string(v - 32))
		} else {
			ns.WriteString(string(v))
		}
	}
	return ns.String()
}

// String 内存地址转换[]byte
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Bytes 内存地址转换string
func Bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			cap int
		}{s, len(s)},
	))
	// x := (*[2]uintptr)(unsafe.Pointer(&s))
	// h := [3]uintptr{x[0], x[1], x[1]}
	// return *(*[]byte)(unsafe.Pointer(&h))
}

// Stamp2Time convert stamp to datetime string
func Stamp2Time(t int64, fmt ...string) string {
	var f string
	if len(fmt) > 0 {
		f = fmt[0]
	} else {
		f = "2006-01-02 15:04:05"
	}
	tm := time.Unix(t, 0)
	return tm.Format(f)
}

// Time2Stamp convert datetime string to stamp
func Time2Stamp(s string) int64 {
	return Time2Stampf(s, "", 8)
}

// Time2Stampf 可根据制定的时间格式和时区转换为当前时区的Unix时间戳
//
//		s：时间字符串
//	 fmt：时间格式
//	 year：2006，month：01，day：02
//	 hour：15，minute：04，second：05
//	 tz：0～12,超范围时使用本地时区
func Time2Stampf(s, fmt string, tz float32) int64 {
	if fmt == "" {
		fmt = "2006-01-02 15:04:05"
	}
	if tz > 12 || tz < 0 {
		_, t := time.Now().Zone()
		tz = float32(t / 3600)
	}
	var loc = time.FixedZone("", int((time.Duration(tz) * time.Hour).Seconds()))
	tm, ex := time.ParseInLocation(fmt, s, loc)
	if ex != nil {
		return 0
	}
	return tm.Unix()
}

// String2Int64 convert string 2 int64
//
//	 args:
//		s: 输入字符串
//		t: 返回数值进制
//	 Return：
//		int64
func String2Int64(s string, t int) int64 {
	x, _ := strconv.ParseInt(s, t, 64)
	return x
}
