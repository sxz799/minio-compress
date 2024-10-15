package utils

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

func CompressFile(file, fileType string, buf *bytes.Buffer) error {
	inputFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	var format imaging.Format
	var options []imaging.EncodeOption
	var decode image.Image
	switch fileType {
	case "image/jpg", "image/jpeg":
		format = imaging.JPEG
		decode, err = jpeg.Decode(inputFile)
		options = append(options, imaging.JPEGQuality(50))
	case "image/png":
		format = imaging.PNG
		decode, err = png.Decode(inputFile)
		options = append(options, imaging.PNGCompressionLevel(png.BestCompression))
	default:
		return fmt.Errorf("不支持的文件类型: %s", fileType) // 处理不支持的文件类型
	}
	if err != nil {
		return err
	}
	err = imaging.Encode(buf, decode, format, options...)
	if err != nil {
		return err
	}
	return nil
}
