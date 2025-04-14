package qrcode

import (
	"image"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCodeImage(data string, size int) image.Image {
	qr, _ := qrcode.New(data, qrcode.High)
	qr.DisableBorder = true
	return qr.Image(size)
}
