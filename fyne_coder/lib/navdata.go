// Package data data
package data

import "fyne.io/fyne/v2"

// NavItem defines the data structure for a tutorial
type NavItem struct {
	Title      string
	Intro      string
	SupportWeb bool
	View       func(w fyne.Window) fyne.CanvasObject
	Object     fyne.CanvasObject
}

var (
	// NavIdx NavIdx
	NavIdx = map[string][]string{
		"":      {"coder", "b64", "url", "timer", "id", "hash", "crypt"},
		"hash":  {"md5", "sha256", "sha512", "hmacsha1", "hmacsha256"},
		"crypt": {"aes128cbc", "aes192cbc", "aes256cbc", "aes128cfb", "aes192cfb", "aes256cfb"},
	}
	// NavItems NavItems
	NavItems = map[string]*NavItem{
		"coder": {
			Title:      "编码",
			Intro:      "字符串编码",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardCoder },
		},
		"b64": {
			Title:      "Base64编码/解码",
			Intro:      "Base64编码/解码",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardB64 },
		},
		"url": {
			Title:      "URL编码/解码",
			Intro:      "URL编码/解码",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardURL },
		},
		"timer": {
			Title:      "时间格式转换",
			Intro:      "时间格式转换",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardTimer },
		},
		"id": {
			Title:      "ID生成",
			Intro:      "生成UUID/ULID",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardID },
		},
		"hash": {
			Title:      "字符串HASH",
			Intro:      "生成字符串的各种HASH值",
			SupportWeb: true,
			Object:     nil,
		},
		"md5": {
			Title:      "MD5",
			Intro:      "生成MD5编码",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardMD5 },
		},
		"sha256": {
			Title:      "SHA256",
			Intro:      "生成SHA256编码",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardSHA256 },
		},
		"sha512": {
			Title:      "SHA512",
			Intro:      "生成SHA512编码",
			SupportWeb: true,
			View:       func(w fyne.Window) fyne.CanvasObject { return cardSHA512 },
		},
		"hmacsha1": {
			Title:      "HMACSHA1",
			Intro:      "生成HMACSHA1编码",
			SupportWeb: true,
			View:       CardHMACSHA1,
		},
		"hmacsha256": {
			Title:      "HMACSHA256",
			Intro:      "生成HMACSHA256编码",
			SupportWeb: true,
			View:       CardHMACSHA256,
		},
		"crypt": {
			Title:      "字符串对称加密",
			Intro:      "字符串对称加密解密",
			SupportWeb: true,
			Object:     nil,
		},
		"aes128cbc": {
			Title:      "AES128CBC",
			Intro:      "AES128CBC加密解密",
			SupportWeb: true,
			View:       CardAES128CBC,
		},
		"aes192cbc": {
			Title:      "AES192CBC",
			Intro:      "AES192CBC加密解密",
			SupportWeb: true,
			View:       CardAES192CBC,
		},
		"aes256cbc": {
			Title:      "AES256CBC",
			Intro:      "AES256CBC加密解密",
			SupportWeb: true,
			View:       CardAES256CBC,
		},
		"aes128cfb": {
			Title:      "AES128CFB",
			Intro:      "AES128CFB加密解密",
			SupportWeb: true,
			View:       CardAES128CFB,
		},
		"aes192cfb": {
			Title:      "AES192CFB",
			Intro:      "AES192CFB加密解密",
			SupportWeb: true,
			View:       CardAES192CFB,
		},
		"aes256cfb": {
			Title:      "AES256CFB",
			Intro:      "AES256CFB加密解密",
			SupportWeb: true,
			View:       CardAES256CFB,
		},
	}
)
