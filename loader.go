package main

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func CheckUrl(fileUrl string) error {
	_, err := url.ParseRequestURI(fileUrl)
	return err
}

func LoadImage(fileUrl, filePath string) error {
	const timeout = 60 * time.Second

	req, err := http.NewRequest("GET", fileUrl, nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: timeout,
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func RemoveImage(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}

func DecodeImage(filePath string) (image.Image, error) {
	imageFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()

	img, _, err := image.Decode(imageFile)
	if err != nil {
		return nil, err
	}

	return img, nil
}
