package data

import (
	layoutx "coder/fyne/layout"
	widgetx "coder/fyne/widget"
	"coder/ulid"
	"encoding/base64"
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	txtIn          = widgetx.MultiLineEntry(15, 0, fyne.TextWrapWord)
	txtOut         = widgetx.MultiLineEntry(15, 0, fyne.TextWrapWord)
	txtIV          = widgetx.MultiLineEntry(1, 0, fyne.TextWrapWord)
	txtKey         = widgetx.MultiLineEntry(1, 0, fyne.TextWrapWord)
	keyLen         = 16
	cw             *CryptoWorker
	btnEn          *widget.Button
	btnDe          *widget.Button
	cardCoder      fyne.CanvasObject
	cardB64        fyne.CanvasObject
	cardURL        fyne.CanvasObject
	cardTimer      fyne.CanvasObject
	cardID         fyne.CanvasObject
	cardMD5        fyne.CanvasObject
	cardSHA256     fyne.CanvasObject
	cardSHA512     fyne.CanvasObject
	cardHMACSHA1   fyne.CanvasObject
	cardHMACSHA256 fyne.CanvasObject
	cardAES        fyne.CanvasObject
)

// InitCards InitCards
func InitCards() {
	txtKey.SetPlaceHolder("设置加密key")
	txtKey.Resize(fyne.NewSize(280, 36))
	txtIV.SetPlaceHolder("设置加密iv")
	txtIV.Resize(fyne.NewSize(220, 36))

	btnEn = widget.NewButtonWithIcon("加密", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		if s := txtKey.Text; len(s) < keyLen {
			for i := len(s); i < keyLen; i++ {
				s += "="
			}
			txtKey.SetText(s)
		}
		if s := txtIV.Text; len(s) < 16 {
			for i := len(s); i < 16; i++ {
				s += "="
			}
			txtIV.SetText(s)
		}
		err := cw.SetKey(txtKey.Text, txtIV.Text)
		if err != nil {
			txtOut.SetText(err.Error())
			return
		}
		txtOut.SetText(cw.Encrypt(txtIn.Text))
	})
	btnDe = widget.NewButtonWithIcon("解密", theme.MoveUpIcon(), func() {
		if txtOut.Text == "" {
			return
		}
		if s := txtKey.Text; len(s) < keyLen {
			for i := len(s); i < keyLen; i++ {
				s += "="
			}
			txtKey.SetText(s)
		}
		if s := txtIV.Text; len(s) < 16 {
			for i := len(s); i < 16; i++ {
				s += "="
			}
			txtIV.SetText(s)
		}
		err := cw.SetKey(txtKey.Text, txtIV.Text)
		if err != nil {
			txtOut.SetText(err.Error())
			return
		}
		txtIn.SetText(cw.Decrypt(txtOut.Text))
	})
	cardCoder = CardCoder(nil)
	cardB64 = CardB64(nil)
	cardURL = CardURL(nil)
	cardTimer = CardTimer(nil)
	cardID = CardID(nil)
	cardMD5 = CardMD5(nil)
	cardSHA256 = CardSHA256(nil)
	cardSHA512 = CardSHA512(nil)
	cardAES = cardCrypt()
}

// CardCoder CardCoder
func CardCoder(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("编码", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		txtOut.SetText(CodeString(txtIn.Text))
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator()), nil, nil, nil, container.NewVSplit(txtIn, txtOut))
}

// CardB64 CardB64
func CardB64(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("Base64编码", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		txtOut.SetText(base64.StdEncoding.EncodeToString([]byte(txtIn.Text)))
	})
	btn2 := widget.NewButtonWithIcon("Base64解码", theme.MoveUpIcon(), func() {
		if txtOut.Text == "" {
			return
		}
		x, err := base64.StdEncoding.DecodeString(txtOut.Text)
		if err != nil {
			txtIn.SetText(err.Error())
		} else {
			txtIn.SetText(string(x))
		}
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator(), btn2), nil, nil, nil, container.NewVSplit(txtIn, txtOut))
}

// CardURL CardURL
func CardURL(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("URL编码", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		txtOut.SetText(url.PathEscape(txtIn.Text))
	})
	btn2 := widget.NewButtonWithIcon("URL解码", theme.MoveUpIcon(), func() {
		if txtOut.Text == "" {
			return
		}
		x, err := url.PathUnescape(txtOut.Text)
		if err != nil {
			txtIn.SetText(err.Error())
		} else {
			txtIn.SetText(x)
		}
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator(), btn2), nil, nil, nil, container.NewVSplit(txtIn, txtOut))
}

// CardTimer CardTimer
func CardTimer(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("时间戳转字符串", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		txtOut.SetText(Stamp2Time(String2Int64(txtIn.Text, 10)))
	})
	btn2 := widget.NewButtonWithIcon("字符串转时间戳", theme.MoveUpIcon(), func() {
		if txtOut.Text == "" {
			return
		}
		txtIn.SetText(fmt.Sprintf("%d", Time2Stamp(txtOut.Text)))
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator(), btn2), nil, nil, nil, container.NewVSplit(txtIn, txtOut))
}

// CardID CardID
func CardID(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("生成UUID", theme.MoveDownIcon(), func() {
		txtOut.SetText(GetUUID1())
	})
	btn2 := widget.NewButtonWithIcon("生成ULID", theme.MoveDownIcon(), func() {
		txtOut.SetText(ulid.Make().String())
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator(), btn2), nil, nil, nil, txtOut)
}

// CardMD5 CardMD5
func CardMD5(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("MD5", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		txtOut.SetText(GetMD5(txtIn.Text))
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator()), nil, nil, nil, container.NewVSplit(txtIn, txtOut))
}

// CardSHA256 CardSHA256
func CardSHA256(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("SHA256", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		w := GetNewCryptoWorker(CryptoSHA256)
		txtOut.SetText(w.Hash([]byte(txtIn.Text)))
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator()), nil, nil, nil, container.NewVSplit(txtIn, txtOut))
}

// CardSHA512 CardSHA512
func CardSHA512(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("SHA512", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		w := GetNewCryptoWorker(CryptoSHA512)
		txtOut.SetText(w.Hash([]byte(txtIn.Text)))
	})
	return container.NewBorder(container.NewHBox(btn1, widget.NewSeparator()), nil, nil, nil, container.NewVSplit(txtIn, txtOut))
}

// CardHMACSHA1 CardHMACSHA1
func CardHMACSHA1(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("HMACSHA1", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		w := GetNewCryptoWorker(CryptoHMACSHA1)
		if txtKey.Text != "" {
			w.SetKey(txtKey.Text, "")
		}
		txtOut.SetText(w.Hash([]byte(txtIn.Text)))
	})
	if fyne.CurrentDevice().IsMobile() {
		return widget.NewForm(
			widget.NewFormItem("设置Key：", txtKey),
			widget.NewFormItem("明文：", txtIn),
			widget.NewFormItem("", btn1),
			widget.NewFormItem("密文", txtOut),
		)
	}
	return container.NewBorder(container.New(&layoutx.HBoxLayout{}, widgetx.RightAlignLabel("设置Key："), txtKey, layout.NewSpacer(), btn1),
		nil,
		nil,
		nil,
		container.NewVSplit(txtIn, txtOut))
}

// CardHMACSHA256 CardHMACSHA256
func CardHMACSHA256(_ fyne.Window) fyne.CanvasObject {
	btn1 := widget.NewButtonWithIcon("HMACSHA256", theme.MoveDownIcon(), func() {
		if txtIn.Text == "" {
			return
		}
		w := GetNewCryptoWorker(CryptoHMACSHA256)
		if txtKey.Text != "" {
			w.SetKey(txtKey.Text, "")
		}
		txtOut.SetText(w.Hash([]byte(txtIn.Text)))
	})
	if fyne.CurrentDevice().IsMobile() {
		return widget.NewForm(
			widget.NewFormItem("设置Key：", txtKey),
			widget.NewFormItem("明文：", txtIn),
			widget.NewFormItem("", btn1),
			widget.NewFormItem("密文", txtOut),
		)
	}
	return container.NewBorder(container.New(&layoutx.HBoxLayout{}, widgetx.RightAlignLabel("设置Key："), txtKey, layout.NewSpacer(), btn1),
		nil,
		nil,
		nil,
		container.NewVSplit(txtIn, txtOut))
}

// CardAES128CBC CardAES128CBC
func CardAES128CBC(_ fyne.Window) fyne.CanvasObject {
	cw = GetNewCryptoWorker(CryptoAES128CBC)
	keyLen = 16
	return cardAES
	// return cardCrypt()
}

// CardAES192CBC CardAES192CBC
func CardAES192CBC(_ fyne.Window) fyne.CanvasObject {
	cw = GetNewCryptoWorker(CryptoAES192CBC)
	keyLen = 24
	return cardAES
	// return cardCrypt()
}

// CardAES256CBC CardAES256CBC
func CardAES256CBC(_ fyne.Window) fyne.CanvasObject {
	cw = GetNewCryptoWorker(CryptoAES256CBC)
	keyLen = 32
	return cardAES
	// return cardCrypt()
}

// CardAES128CFB CardAES128CFB
func CardAES128CFB(_ fyne.Window) fyne.CanvasObject {
	cw = GetNewCryptoWorker(CryptoAES128CFB)
	keyLen = 16
	return cardAES
	// return cardCrypt()
}

// CardAES192CFB CardAES192CFB
func CardAES192CFB(_ fyne.Window) fyne.CanvasObject {
	cw = GetNewCryptoWorker(CryptoAES192CFB)
	keyLen = 24
	return cardAES
	// return cardCrypt()
}

// CardAES256CFB CardAES256CFB
func CardAES256CFB(_ fyne.Window) fyne.CanvasObject {
	cw = GetNewCryptoWorker(CryptoAES256CFB)
	keyLen = 32
	return cardAES
	// return cardCrypt()
}

func cardCrypt() fyne.CanvasObject {
	if fyne.CurrentDevice().IsMobile() {
		return widget.NewForm(
			widget.NewFormItem("设置Key：", txtKey),
			widget.NewFormItem("设置IV：", txtIV),
			widget.NewFormItem("明文：", txtIn),
			widget.NewFormItem("", container.NewHBox(btnEn, btnDe)),
			widget.NewFormItem("密文", txtOut),
		)
	}
	return container.NewBorder(container.New(&layoutx.HBoxLayout{},
		widgetx.RightAlignLabel("设置Key："), txtKey,
		widgetx.RightAlignLabel("设置IV："), txtIV,
		btnEn, btnDe),
		nil,
		nil,
		nil,
		container.NewVSplit(txtIn, txtOut))
}
