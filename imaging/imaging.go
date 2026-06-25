// Package imaging 提供图像处理工具函数
package imaging

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// LoadImage 加载图像文件 (支持 jpg, png, bmp, tif, gif, webp)
func LoadImage(path string) (image.Image, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("读取文件失败: %w", err)
	}
	return DecodeImage(data, detectFormatByPath(path))
}

// DecodeImage 从字节数据解码图像
func DecodeImage(data []byte, format string) (image.Image, string, error) {
	reader := bytes.NewReader(data)

	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		img, err := jpeg.Decode(reader)
		return img, "jpeg", err
	case "png":
		img, err := png.Decode(reader)
		return img, "png", err
	case "gif":
		img, err := gif.Decode(reader)
		return img, "gif", err
	case "bmp":
		img, err := bmp.Decode(reader)
		return img, "bmp", err
	case "tif", "tiff":
		img, err := tiff.Decode(reader)
		return img, "tiff", err
	case "webp":
		img, err := webp.Decode(reader)
		return img, "webp", err
	default:
		// 尝试自动检测
		img, fmt, err := image.Decode(reader)
		return img, fmt, err
	}
}

// EncodeToJPEG 将图像编码为 JPEG 字节
func EncodeToJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeToPNG 将图像编码为 PNG 字节
func EncodeToPNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeToBMP 将图像编码为 BMP 字节
func EncodeToBMP(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeToTIFF 将图像编码为 TIFF 字节
func EncodeToTIFF(img image.Image, compression string) ([]byte, error) {
	var buf bytes.Buffer
	var comp tiff.CompressionType
	switch compression {
	case "lzw":
		comp = tiff.LZW
	case "deflate":
		comp = tiff.Deflate
	default:
		comp = tiff.Uncompressed
	}
	if err := tiff.Encode(&buf, img, &tiff.Options{Compression: comp}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeImage 按指定格式编码图像
func EncodeImage(img image.Image, format string, quality int, tiffCompression string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		return EncodeToJPEG(img, quality)
	case "png":
		return EncodeToPNG(img)
	case "bmp":
		return EncodeToBMP(img)
	case "tif", "tiff":
		return EncodeToTIFF(img, tiffCompression)
	default:
		return EncodeToJPEG(img, quality)
	}
}

// Clip 裁剪图像
func Clip(img image.Image, x, y, w, h int) (image.Image, error) {
	bounds := img.Bounds()
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > bounds.Max.X {
		w = bounds.Max.X - x
	}
	if y+h > bounds.Max.Y {
		h = bounds.Max.Y - y
	}
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid clip dimensions")
	}

	rect := image.Rect(x, y, x+w, y+h)
	result := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(result, result.Bounds(), img, rect.Min, draw.Src)
	return result, nil
}

// Thumbnail 生成缩略图（保持宽高比，最大宽度 maxWidth）
func Thumbnail(img image.Image, maxWidth int) (image.Image, error) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxWidth {
		return img, nil
	}

	newW := maxWidth
	newH := h * maxWidth / w

	result := image.NewRGBA(image.Rect(0, 0, newW, newH))
	// 简单近邻采样缩放
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := x * w / newW
			srcY := y * h / newH
			result.Set(x, y, img.At(srcX, srcY))
		}
	}
	return result, nil
}

// MergeHorz 水平合并图像
func MergeHorz(imgs []image.Image, align string, interval int) (image.Image, error) {
	if len(imgs) == 0 {
		return nil, fmt.Errorf("no images to merge")
	}
	if len(imgs) == 1 {
		return imgs[0], nil
	}

	// 计算总宽度和最大高度
	totalW := (len(imgs) - 1) * interval
	maxH := 0
	for _, img := range imgs {
		b := img.Bounds()
		totalW += b.Dx()
		if b.Dy() > maxH {
			maxH = b.Dy()
		}
	}

	result := image.NewRGBA(image.Rect(0, 0, totalW, maxH))
	// 白底填充
	draw.Draw(result, result.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	x := 0
	for _, img := range imgs {
		b := img.Bounds()
		y := 0
		switch align {
		case "bottom":
			y = maxH - b.Dy()
		case "center":
			y = (maxH - b.Dy()) / 2
		default: // top
			y = 0
		}
		draw.Draw(result, image.Rect(x, y, x+b.Dx(), y+b.Dy()), img, b.Min, draw.Src)
		x += b.Dx() + interval
	}

	return result, nil
}

// MergeVert 垂直合并图像
func MergeVert(imgs []image.Image, align string, interval int) (image.Image, error) {
	if len(imgs) == 0 {
		return nil, fmt.Errorf("no images to merge")
	}
	if len(imgs) == 1 {
		return imgs[0], nil
	}

	// 计算总高度和最大宽度
	totalH := (len(imgs) - 1) * interval
	maxW := 0
	for _, img := range imgs {
		b := img.Bounds()
		totalH += b.Dy()
		if b.Dx() > maxW {
			maxW = b.Dx()
		}
	}

	result := image.NewRGBA(image.Rect(0, 0, maxW, totalH))
	// 白底填充
	draw.Draw(result, result.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	y := 0
	for _, img := range imgs {
		b := img.Bounds()
		x := 0
		switch align {
		case "right":
			x = maxW - b.Dx()
		case "center":
			x = (maxW - b.Dx()) / 2
		default: // left
			x = 0
		}
		draw.Draw(result, image.Rect(x, y, x+b.Dx(), y+b.Dy()), img, b.Min, draw.Src)
		y += b.Dy() + interval
	}

	return result, nil
}

// SplitHorz 水平分割图像（在 location 处切分）
func SplitHorz(img image.Image, location int) (image.Image, image.Image, error) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if location <= 0 || location >= w {
		return nil, nil, fmt.Errorf("invalid split location")
	}

	left := image.NewRGBA(image.Rect(0, 0, location, h))
	right := image.NewRGBA(image.Rect(0, 0, w-location, h))

	draw.Draw(left, left.Bounds(), img, bounds.Min, draw.Src)
	draw.Draw(right, right.Bounds(), img, image.Point{bounds.Min.X + location, bounds.Min.Y}, draw.Src)

	return left, right, nil
}

// SplitVert 垂直分割图像（在 location 处切分）
func SplitVert(img image.Image, location int) (image.Image, image.Image, error) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if location <= 0 || location >= h {
		return nil, nil, fmt.Errorf("invalid split location")
	}

	top := image.NewRGBA(image.Rect(0, 0, w, location))
	bottom := image.NewRGBA(image.Rect(0, 0, w, h-location))

	draw.Draw(top, top.Bounds(), img, bounds.Min, draw.Src)
	draw.Draw(bottom, bottom.Bounds(), img, image.Point{bounds.Min.X, bounds.Min.Y + location}, draw.Src)

	return top, bottom, nil
}

// MakeMultiPageTIFF 创建多页 TIFF
func MakeMultiPageTIFF(imgs []image.Image) ([]byte, error) {
	// 简化实现：将多张图片合并为一张垂直长图
	result, err := MergeVert(imgs, "center", 0)
	if err != nil {
		return nil, err
	}
	return EncodeToTIFF(result, "none")
}

// MakeZipFile 创建 ZIP 压缩文件
func MakeZipFile(imagesData [][]byte, names []string) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for i, data := range imagesData {
		name := fmt.Sprintf("image_%03d.jpg", i)
		if i < len(names) {
			name = names[i]
		}
		f, err := w.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// detectFormatByPath 从文件路径检测格式
func detectFormatByPath(path string) string {
	path = strings.ToLower(path)
	if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		return "jpg"
	}
	if strings.HasSuffix(path, ".png") {
		return "png"
	}
	if strings.HasSuffix(path, ".bmp") {
		return "bmp"
	}
	if strings.HasSuffix(path, ".tif") || strings.HasSuffix(path, ".tiff") {
		return "tiff"
	}
	if strings.HasSuffix(path, ".gif") {
		return "gif"
	}
	if strings.HasSuffix(path, ".webp") {
		return "webp"
	}
	return ""
}

// ImageToBase64 将图像编码为 JPEG 后返回 base64 字符串
func ImageToBase64(img image.Image) (string, error) {
	data, err := EncodeToJPEG(img, 80)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// ImageDataToBase64 将原始图像数据转为 base64
func ImageDataToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
