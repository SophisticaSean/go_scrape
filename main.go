package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
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

func main() {
	thread_url := ""
	img_list, title := digest4chanPage(thread_url)
	fmt.Println(img_list)
	re := regexp.MustCompile(`\d{10,30}\.\w*`)
	for _, item := range img_list {
		download_dir := strings.Replace(title, " ", "_", -1)
		filename := re.FindString(item)
		fmt.Println(filename, download_dir)
		//go
	}
}
