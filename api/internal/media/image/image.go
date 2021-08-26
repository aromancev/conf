package image

import (
	"fmt"
	"image"
	"os"
	"path"

	"github.com/sunshineplan/imgconv"
)

type Converter struct {
	mediaDir string
}

func NewConverter(mediaDir string) *Converter {
	return &Converter{mediaDir: mediaDir}
}

func (p *Converter) Convert(mediaID string) error {
	file, err := os.Open(path.Join(p.mediaDir, mediaID, "img.raw"))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	src, err := imgconv.Decode(file)

	err = resizeAndSave(src, path.Join(p.mediaDir, mediaID, "256.png"), imgconv.ResizeOption{Width: 256, Height: 256})
	if err != nil {
		return fmt.Errorf("failed to convert image 256x256: %w", err)
	}

	err = resizeAndSave(src, path.Join(p.mediaDir, mediaID, "128.png"), imgconv.ResizeOption{Width: 128, Height: 128})
	if err != nil {
		return fmt.Errorf("failed to convert image 128x128: %w", err)
	}

	err = resizeAndSave(src, path.Join(p.mediaDir, mediaID, "64.png"), imgconv.ResizeOption{Width: 64, Height: 64})
	if err != nil {
		return fmt.Errorf("failed to convert image 64x64: %w", err)
	}

	_ = os.Remove(path.Join(p.mediaDir, mediaID, "img.raw"))
	return nil
}

func resizeAndSave(src image.Image, filePath string, ro imgconv.ResizeOption) error {
	dstImage := imgconv.Resize(src, ro)

	fileImg, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	err = imgconv.Write(fileImg, dstImage, imgconv.FormatOption{Format: imgconv.PNG})
	if err != nil {
		return err
	}

	return nil
}
