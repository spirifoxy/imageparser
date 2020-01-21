package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"time"
)

const downloadPath = "./tmp"

type resultData struct {
	url       string
	hexColors []string
}

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
	traceout   = flag.String("traceout", "", "write trace output to `file`")
)

func createDownloadFolder() error {
	if _, err := os.Stat(downloadPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(downloadPath); err != nil {
			fmt.Println(err)
		}
	}

	if err := os.Mkdir(downloadPath, os.ModeDir); err != nil {
		return err
	}

	return nil
}

func main() {
	const (
		inputFilename   = "./input.txt"
		resultsFilename = "results.csv"
		workerCount     = 10
	)

	flag.Parse()

	if *traceout != "" {
		f, err := os.Create(*traceout)
		if err != nil {
			log.Fatal("could not create trace: ", err)
		}
		defer f.Close()
		if err := trace.Start(f); err != nil {
			log.Fatal("could not start trace: ", err)
		}
		defer trace.Stop()
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if err := createDownloadFolder(); err != nil {
		log.Fatal(err)
	}

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	urls := make(chan string)
	results := make(chan resultData)
	done := make(chan bool)

	go parseFile(inputFile, urls)
	go resultToCsv(resultsFilename, results, done)
	createWorkerPool(workerCount, urls, results)
	<-done

	if err := os.RemoveAll(downloadPath); err != nil {
		fmt.Println(err)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func parseFile(inputFile *os.File, urls chan string) {
	sc := bufio.NewScanner(inputFile)
	for sc.Scan() {
		imgUrl := sc.Text()
		if err := CheckUrl(imgUrl); err != nil {
			fmt.Println(err)
			continue
		}
		urls <- imgUrl
	}

	if err := sc.Err(); err != nil {
		fmt.Println(err)
	}
	close(urls)
}

func resultToCsv(filename string, results chan resultData, done chan bool) {
	resultsFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer resultsFile.Close()

	csvWriter := csv.NewWriter(resultsFile)
	defer csvWriter.Flush()

	for result := range results {
		line := append([]string{result.url}, result.hexColors...)
		if err := csvWriter.Write(line); err != nil {
			fmt.Println(err)
		}
	}
	done <- true
}

func createWorkerPool(workersCount int, urls chan string, results chan resultData) {
	var wg sync.WaitGroup
	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go startWorker(&wg, urls, results)
	}
	wg.Wait()
	close(results)
}

func startWorker(wg *sync.WaitGroup, urls chan string, results chan resultData) {
	for url := range urls {
		res, err := processUrl(url)
		if err != nil {
			continue
		}
		results <- *res
	}
	wg.Done()
}

func processUrl(fileUrl string) (*resultData, error) {
	const reqResultsNum = 3

	var filename string
	segments := strings.Split(fileUrl, "/")
	if len(segments) > 0 {
		filename = segments[len(segments)-1]
	}
	// to definitely avoid collisions between filenames in case of processing the same urls
	filePath := fmt.Sprintf("%s/%d_%d_%s.jpg", downloadPath, time.Now().UnixNano(), rand.Int(), filename)

	if err := LoadImage(fileUrl, filePath); err != nil {
		return nil, fmt.Errorf("%s: %s", fileUrl, err.Error())
	}

	img, err := DecodeImage(filePath)
	if err != nil {
		fmt.Println(fileUrl + ": " + err.Error())
		runtime.Goexit()
		return nil, fmt.Errorf("%s: %s", fileUrl, err.Error())
	}

	hex := Kmeans(img, reqResultsNum)
	RemoveImage(filePath) // ignore errors, the whole folder will be deleted anyway

	return &resultData{
		url:       fileUrl,
		hexColors: hex,
	}, nil
}
