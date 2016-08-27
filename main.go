package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
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

func digest4chanPage(url string) (img_list []string, title string) {
	resp := getPage(url)
	body := html.NewTokenizer(resp.Body)

	// get the root post info
	//title := ""
	for {
		element := body.Next()

		switch {
		case element == html.ErrorToken:
			return
		case element == html.StartTagToken:
			element := body.Token()

			isSpan := element.Data == "span"
			if isSpan {
				for _, a := range element.Attr {
					if a.Key == "class" {
						switch {
						case a.Val == "subject":
							// advance forward once into data
							body.Next()
							element = body.Token()

							title = element.Data
							//fmt.Println(element.Data)
						}
					}
				}
			}

			isDiv := element.Data == "a"
			if isDiv {
				for _, a := range element.Attr {
					if a.Key == "class" {
						switch {
						case a.Val == "fileThumb":
							for _, b := range element.Attr {
								if b.Key == "href" {
									img_url := strings.Replace(b.Val, "//", "http://", 1)
									img_list = append(img_list, img_url)
									//fmt.Println(reflect.TypeOf(b.Val))
								}
							}
						}
					}
				}
			}
		}
	}
	return img_list, title
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
	thread_url := os.Args[1]

	img_list, title := digest4chanPage(thread_url)

	re := regexp.MustCompile(`\d{10,30}\.\w*`)

	download_dir := safeStringToFilepath(title)
	os.MkdirAll(download_dir, os.FileMode(0777))

	var wg sync.WaitGroup
	wg.Add(len(img_list))

	for _, item := range img_list {
		fmt.Println("hello")
		go func(item string) {
			filename := re.FindString(item)
			filepath := download_dir + "/" + filename
			downloadFile(filepath, item)
			wg.Done()
		}(item)
		//time.Sleep(100 * time.Millisecond)
	}
	wg.Wait()
	fmt.Println(len(img_list))
	//fmt.Println(os.Args[1])
}
