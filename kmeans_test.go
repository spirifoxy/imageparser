package main

import (
	"image/color"
	"testing"
)

func TestCreateColorData(t *testing.T) {
	data := createColorData(color.White, 0)

	if data.r != uint32(255) || data.g != uint32(255) || data.b != uint32(255) {
		t.Errorf("Wrong color to rgb convertion: %d,%d,%d; expected: 255,255,255.", data.r, data.g, data.b)
	}

	if data.hexColor != "#FFFFFF" {
		t.Errorf("Wrong color to hex convertion: %s; expected: #FFFFFF.", data.hexColor)
	}
}

func TestKmeansRed(t *testing.T) {
	// 250,0,0
	img, err := DecodeImage("./test_red.jpg")
	if err != nil {
		t.Error(err)
	}

	type testCase struct {
		reqResultsNum int
		expectedHex   []string
	}

	cases := []testCase{
		{
			reqResultsNum: 1,
			expectedHex:   []string{"#FA0000"},
		},
		{
			reqResultsNum: 3,
			expectedHex:   []string{"#FA0000", "#000000", "#000000"},
		},
	}

	for _, c := range cases {
		resultHex := Kmeans(img, c.reqResultsNum)
		if len(resultHex) != c.reqResultsNum {
			t.Errorf("Wrong result colors number: %d; expected: %d.", len(resultHex), c.reqResultsNum)
		}
		for i := 0; i < len(resultHex); i++ {
			if resultHex[i] != c.expectedHex[i] {
				t.Errorf("Wrong result color value: %s; expected: %s.", resultHex[i], c.expectedHex[i])
			}
		}
	}
}
