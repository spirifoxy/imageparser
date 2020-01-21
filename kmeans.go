package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"time"
)

type ColorData struct {
	r, g, b  uint32
	hexColor string
	count    int
}

type centroid struct {
	index int
	color ColorData
}

func createColorData(color color.Color, count int) ColorData {
	r, g, b, _ := color.RGBA()
	r, g, b = r>>8, g>>8, b>>8
	return ColorData{
		r:        r,
		g:        g,
		b:        b,
		hexColor: fmt.Sprintf("#%02X%02X%02X", r, g, b),
		count:    count,
	}
}

func (c *ColorData) String() string {
	return c.hexColor
}

func (c *ColorData) rgb() (uint32, uint32, uint32) {
	return c.r, c.g, c.b
}

func (c *ColorData) increaseCount() {
	c.count++
}

func Kmeans(img image.Image, resultsNum int) []string {
	rand.Seed(time.Now().UTC().UnixNano())

	colors := countColors(img)

	centroids := initializeCentroids(resultsNum, colors)
	clusters := initializeClusters(resultsNum)
	clusters[0] = colors

	var permutationsFinished bool
	for permutationsFinished == false {
		permutationsFinished = true

		tClusters := initializeClusters(resultsNum)

		for i := 0; i < resultsNum; i++ {
			for _, colorItem := range clusters[i] {
				var newIndex int
				closestCentroid, err := findClosestCentroid(colorItem, centroids)
				if err != nil {
					// is not supposed to happen at all, but just in case
					newIndex = i
				} else {
					newIndex = closestCentroid.index
				}
				if newIndex != i {
					permutationsFinished = false
				}
				tClusters[newIndex] = append(tClusters[newIndex], colorItem)
			}
		}
		clusters = tClusters
		centroids = createMedianCentroids(clusters)
	}

	hexColors := make([]string, resultsNum)
	for i := 0; i < len(centroids); i++ {
		hexColors[i] = centroids[i].color.String()
	}
	return hexColors
}

func initializeCentroids(count int, colors []ColorData) []centroid {
	centroids := make([]centroid, count)
	for i := 0; i < len(centroids); i++ {
		centroids[i].index = i
	}

	isChosen := make(map[int]bool)
	for i := 0; i < count; {
		randomIndex := rand.Intn(len(colors))

		if isChosen[randomIndex] == true {
			continue
		}

		isChosen[randomIndex] = true
		centroids[i].color = colors[randomIndex]

		// in case we took all the colors
		if len(isChosen) == len(colors) {
			break
		}
		i++
	}

	return centroids
}

func initializeClusters(count int) [][]ColorData {
	clusters := make([][]ColorData, count)
	for i := 0; i < len(clusters); i++ {
		clusters[i] = []ColorData{}
	}
	return clusters
}

func countColors(img image.Image) []ColorData {
	colorsCount := make(map[string]ColorData)
	imageBounds := img.Bounds()

	for i := 0; i < imageBounds.Max.X; i++ {
		for j := 0; j < imageBounds.Max.Y; j++ {
			colorAt := img.At(i, j)
			newColorData := createColorData(colorAt, 0)
			if existColorData, ok := colorsCount[newColorData.hexColor]; ok {
				existColorData.increaseCount()
				colorsCount[existColorData.hexColor] = existColorData
			} else {
				newColorData.increaseCount()
				colorsCount[newColorData.hexColor] = newColorData
			}
		}
	}

	colorsArr := make([]ColorData, len(colorsCount))
	var index int
	for _, colorItem := range colorsCount {
		colorsArr[index] = colorItem
		index++
	}
	return colorsArr
}

func findClosestCentroid(colorItem ColorData, centroids []centroid) (*centroid, error) {

	var closestCentroid *centroid = nil
	var closestDistance = math.MaxInt32
	for i := 0; i < len(centroids); i++ {
		centR, centG, centB := centroids[i].color.rgb()
		distance := int((centR-colorItem.r)*(centR-colorItem.r) + (centG-colorItem.g)*(centG-colorItem.g) + (centB-colorItem.b)*(centB-colorItem.b))
		if distance < closestDistance {
			closestDistance = distance
			closestCentroid = &centroids[i]
		}
	}

	if closestCentroid == nil {
		return nil, fmt.Errorf("closest centroid is not found")
	}

	return closestCentroid, nil
}

func createMedianCentroids(clusters [][]ColorData) []centroid {
	var centroids []centroid

	for i := 0; i < len(clusters); i++ {

		var r, g, b uint32
		var rs, gs, bs []uint32
		var count int
		for _, colorItem := range clusters[i] {
			count += colorItem.count
			rs = append(rs, colorItem.r)
			gs = append(gs, colorItem.g)
			bs = append(bs, colorItem.b)
		}

		sort.Slice(rs, func(i, j int) bool { return rs[i] < rs[j] })
		sort.Slice(gs, func(i, j int) bool { return gs[i] < gs[j] })
		sort.Slice(bs, func(i, j int) bool { return bs[i] < bs[j] })

		if len(rs) > 0 {
			r = rs[len(rs)/2]
		}

		if len(gs) > 0 {
			g = gs[len(gs)/2]
		}

		if len(bs) > 0 {
			b = bs[len(bs)/2]
		}

		centroid := centroid{
			index: i,
			color: createColorData(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0}, count),
		}

		centroids = append(centroids, centroid)
	}

	return centroids
}
