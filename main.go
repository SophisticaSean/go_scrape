package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func downloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getPage(url string) (resp *http.Response) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Website is giving us issues, retrying.")
		time.Sleep(2 * time.Second)
		resp = getPage(url)
	}
	return
}

func digest4chanPage(url string) (imgList []string, title string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}
	doc.Find("span.subject").Each(func(i int, s *goquery.Selection) {
		fmt.Println(s.Text())
		title = s.Text()
	})
	doc.Find("div.fileText a").Each(func(i int, s *goquery.Selection) {
		imgURL, _ := s.Attr("href")
		imgURL = strings.Replace(imgURL, "//", "http://", 100)
		imgList = append(imgList, imgURL)
	})
	return imgList, title
}

func safeStringToFilepath(badString string) (filepath string) {
	badChars := []string{
		"'",
		`"`,
		`\`,
		`/`,
		"#",
		"!",
		"?",
		"(",
		")",
	}

	for _, badChar := range badChars {
		badString = strings.Replace(badString, badChar, "", -1)
	}
	filepath = strings.Replace(badString, " ", "_", -1)
	return
}

//func verboseWait

func main() {
	threadURL := os.Args[1]

	preImgList, title := digest4chanPage(threadURL)
	imgList := []string{}

	re := regexp.MustCompile(`\d{10,30}\.\w*`)

	downloadDir := safeStringToFilepath(title)
	os.MkdirAll(downloadDir, os.FileMode(0777))
	files, _ := ioutil.ReadDir(downloadDir)

	// Skip files we've already downloaded
	for _, item := range preImgList {
		skip := false
		for _, file := range files {
			if strings.Contains(item, file.Name()) {
				skip = true
			}
		}
		if !skip {
			imgList = append(imgList, item)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(imgList))

	for _, item := range imgList {
		go func(item string) {
			filename := re.FindString(item)
			filepath := downloadDir + "/" + filename
			downloadFile(filepath, item)
			wg.Done()
		}(item)
		//time.Sleep(100 * time.Millisecond)
	}
	wg.Wait()
	fmt.Println(len(imgList))
	//fmt.Println(os.Args[1])
}
