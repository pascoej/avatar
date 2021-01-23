package avatar

import (
	"errors"
	"github.com/lucasb-eyer/go-colorful"
	"gopkg.in/go-playground/colors.v1"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"regexp"
	"strconv"
)

func decodePng(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

var hrRegex = regexp.MustCompile(`hr\((.+)\)`)

func transformHue(img image.Image, rotation float64) (image.Image, error) {
	result := image.NewRGBA(img.Bounds())
	colorMap := make(map[color.Color]color.Color)
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			xyc := img.At(x, y)
			if nc, ok := colorMap[xyc]; ok {
				result.Set(x, y, nc)
			} else {
				_, _, _, a := xyc.RGBA()
				if a > 0 {
					if cc, ok := colorful.MakeColor(xyc); ok {
						hue, sat, lum := cc.Hcl()
						var newHue = hue + rotation
						r, g, b, _ := colorful.Hcl(newHue, sat, lum).RGBA()
						newColor := color.RGBA{
							R: uint8(r),
							G: uint8(g),
							B: uint8(b),
							A: uint8(a),
						}
						colorMap[xyc] = newColor
						result.Set(x, y, newColor)
					} else {
						return nil, errors.New("could not make color")
					}
				} else {
					colorMap[xyc] = xyc
					result.Set(x, y, xyc)
				}
			}
		}
	}
	return result, nil
}

func transformColor(img image.Image, c *colors.RGBColor) image.Image {
	result := image.NewRGBA(img.Bounds())
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			xyc := img.At(x, y)
			_, _, _, a := xyc.RGBA()
			if a > 0 {
				result.Set(x, y, color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(a)})
			}
		}
	}
	return result
}

func tof(in string) float64 {
	if f, err := strconv.ParseFloat(in, 64); err == nil {
		return f
	}
	return 0.0
}

func transform(img image.Image, param string) (image.Image, error) {
	if param == "" {
		return img, nil
	}
	if res := hrRegex.FindStringSubmatch(param); len(res) == 2 {
		return transformHue(img, tof(res[1]))
	}
	if res, err := colors.ParseHEX(param); err == nil {
		return transformColor(img, res.ToRGB()), nil
	}
	if res, err := colors.ParseRGB(param); err == nil {
		return transformColor(img, res.ToRGB()), nil
	}
	return nil, errors.New("bad param")
}

func Create(hairParam string, shirtParam string) (image.Image, error) {
	shirt, err := decodePng("shirt.png")
	if err != nil {
		return nil, err
	}
	result := image.NewRGBA(shirt.Bounds())
	changedShirt, err := transform(shirt, shirtParam)
	if err != nil {
		return nil, err
	}
	draw.Draw(result, changedShirt.Bounds(), changedShirt, image.Point{}, draw.Over)

	skin, err := decodePng("skin.png")
	if err != nil {
		return nil, err
	}
	draw.Draw(result, skin.Bounds(), skin, image.Point{}, draw.Over)

	hair, err := decodePng("hair.png")
	if err != nil {
		return nil, err
	}
	changedHair, err := transform(hair, hairParam)
	if err != nil {
		return nil, err
	}
	draw.Draw(result, changedHair.Bounds(), changedHair, image.Point{}, draw.Over)

	shading, err := decodePng("shading.png")
	if err != nil {
		return nil, err
	}
	draw.Draw(result, shading.Bounds(), shading, image.Point{}, draw.Over)
	outline, err := decodePng("outline.png")
	if err != nil {
		return nil, err
	}
	draw.Draw(result, outline.Bounds(), outline, image.Point{}, draw.Over)
	return result, nil
}
