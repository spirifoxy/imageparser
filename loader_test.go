package main

import (
	"os"
	"testing"
)

func TestCheckUrl(t *testing.T) {
	url := "https://google.com"
	if err := CheckUrl(url); err != nil {
		t.Error(err)
	}
}

func TestRemoveImage(t *testing.T) {
	path := "./test_image.jpg"
	file, err := os.Create(path)
	if err != nil {
		t.Error(err)
	}
	file.Close()

	// try to remove test file in case RemoveImage fails
	defer os.Remove(path)

	if err := RemoveImage(path); err != nil {
		t.Error(err)
	}
}

func TestLoadImage(t *testing.T) {
	url := "https://i.imgur.com/FApqk3D.jpg"
	path := "./test_image.jpg"
	if err := LoadImage(url, path); err != nil {
		t.Error(err)
	}
	os.Remove(path)
}

func TestDecodeImage(t *testing.T) {
	url := "https://i.imgur.com/FApqk3D.jpg"
	path := "./test_image.jpg"
	if err := LoadImage(url, path); err != nil {
		t.Error(err)
	}
	defer os.Remove(path)

	_, err := DecodeImage(path)
	if err != nil {
		t.Error(err)
	}
}
