package web

import (
	"avatar/avatar"
	"bytes"
	"fmt"
	"github.com/nfnt/resize"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
	"image"
	"image/draw"
	"image/png"
	"net/http"
	"strconv"
	"time"
)

const (
	Megabyte = 1 << 20
	MaxWidth = 2000
)

func avatarHandler(w http.ResponseWriter, r *http.Request) {
	shirt := r.URL.Query().Get("shirt")
	hair := r.URL.Query().Get("hair")
	var resultImage image.Image
	if img, err := avatar.Create(hair, shirt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		resultImage = img
	}

	if width := r.URL.Query().Get("width"); width != "" {
		if widthI, err := strconv.ParseInt(width, 10, 32); err == nil && widthI > 0 && widthI < MaxWidth {
			resultImage = resize.Resize(uint(widthI), 0, resultImage, resize.Bicubic)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Convenient for some circle avatar pickers, expands canvas width to height
	if square := r.URL.Query().Get("square"); square == "true" {
		height := resultImage.Bounds().Dy()
		squared := image.NewRGBA(image.Rect(0, 0, height, height))
		sx := (height - resultImage.Bounds().Dx()) / 2
		bounds := image.Rect(sx, 0, sx+resultImage.Bounds().Dx(), height)
		draw.Draw(squared, bounds, resultImage, image.Point{}, draw.Over)
		resultImage = squared
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, resultImage); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Listen(addr string) error {
	memcache, err := memory.NewAdapter(
		memory.AdapterWithCapacity(16*Megabyte),
		memory.AdapterWithAlgorithm(memory.LRU),
	)
	if err != nil {
		return err
	}
	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcache),
		cache.ClientWithTTL(10*time.Minute),
	)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Listening on %s", addr))
	return http.ListenAndServe(
		fmt.Sprintf(addr),
		cacheClient.Middleware(http.HandlerFunc(avatarHandler)),
	)
}
