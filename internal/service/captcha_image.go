package service

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/rand"
)

// generateCaptchaImage 生成简单的数字验证码图片，返回 base64 字符串
func generateCaptchaImage(code string) string {
	width, height := 120, 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 背景色
	bgColor := color.RGBA{R: 240, G: 240, B: 245, A: 255}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, bgColor)
		}
	}

	// 干扰线
	lineColors := []color.RGBA{
		{R: 200, G: 200, B: 200, A: 255},
		{R: 180, G: 180, B: 220, A: 255},
		{R: 220, G: 180, B: 180, A: 255},
	}
	for i := 0; i < 4; i++ {
		c := lineColors[i%len(lineColors)]
		y := rand.Intn(height)
		for x := 0; x < width; x++ {
			img.Set(x, y+rand.Intn(3)-1, c)
		}
	}

	// 绘制数字（使用简单像素字体）
	textColor := color.RGBA{R: 50, G: 50, B: 100, A: 255}
	charWidth := width / (len(code) + 1)
	for i, ch := range code {
		drawDigit(img, int(ch-'0'), charWidth*(i+1)-10, 8, textColor)
	}

	// 噪点
	for i := 0; i < 100; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		img.Set(x, y, color.RGBA{
			R: uint8(rand.Intn(200)),
			G: uint8(rand.Intn(200)),
			B: uint8(rand.Intn(200)),
			A: 255,
		})
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// 简单 5x7 像素字体
var digitPatterns = [10][7][5]bool{
	{ // 0
		{false, true, true, true, false},
		{true, false, false, false, true},
		{true, false, false, true, true},
		{true, false, true, false, true},
		{true, true, false, false, true},
		{true, false, false, false, true},
		{false, true, true, true, false},
	},
	{ // 1
		{false, false, true, false, false},
		{false, true, true, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, true, true, true, false},
	},
	{ // 2
		{false, true, true, true, false},
		{true, false, false, false, true},
		{false, false, false, false, true},
		{false, false, true, true, false},
		{false, true, false, false, false},
		{true, false, false, false, false},
		{true, true, true, true, true},
	},
	{ // 3
		{false, true, true, true, false},
		{true, false, false, false, true},
		{false, false, false, false, true},
		{false, false, true, true, false},
		{false, false, false, false, true},
		{true, false, false, false, true},
		{false, true, true, true, false},
	},
	{ // 4
		{false, false, false, true, false},
		{false, false, true, true, false},
		{false, true, false, true, false},
		{true, false, false, true, false},
		{true, true, true, true, true},
		{false, false, false, true, false},
		{false, false, false, true, false},
	},
	{ // 5
		{true, true, true, true, true},
		{true, false, false, false, false},
		{true, true, true, true, false},
		{false, false, false, false, true},
		{false, false, false, false, true},
		{true, false, false, false, true},
		{false, true, true, true, false},
	},
	{ // 6
		{false, true, true, true, false},
		{true, false, false, false, false},
		{true, false, false, false, false},
		{true, true, true, true, false},
		{true, false, false, false, true},
		{true, false, false, false, true},
		{false, true, true, true, false},
	},
	{ // 7
		{true, true, true, true, true},
		{false, false, false, false, true},
		{false, false, false, true, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
		{false, false, true, false, false},
	},
	{ // 8
		{false, true, true, true, false},
		{true, false, false, false, true},
		{true, false, false, false, true},
		{false, true, true, true, false},
		{true, false, false, false, true},
		{true, false, false, false, true},
		{false, true, true, true, false},
	},
	{ // 9
		{false, true, true, true, false},
		{true, false, false, false, true},
		{true, false, false, false, true},
		{false, true, true, true, true},
		{false, false, false, false, true},
		{false, false, false, false, true},
		{false, true, true, true, false},
	},
}

func drawDigit(img *image.RGBA, digit, x, y int, c color.RGBA) {
	if digit < 0 || digit > 9 {
		return
	}
	scale := 3
	pattern := digitPatterns[digit]
	for row := 0; row < 7; row++ {
		for col := 0; col < 5; col++ {
			if pattern[row][col] {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						img.Set(x+col*scale+dx, y+row*scale+dy, c)
					}
				}
			}
		}
	}
}
