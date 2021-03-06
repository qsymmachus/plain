package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/common/log"
)

const (
	DefaultURL   = "https://en.wikipedia.org/wiki/%22Hello,_World!%22_program"
	TextSelector = "p, h1, h2, h3, h4, h5, h6"
)

// Retrieves the document at the URL specified by the '-url' flag, and prints a
// plaintext representation of its content to standard output. For example:
//
//   plain -url http://example.com
//
// Optionally you can output the text to a file instead using the '-file' flag:
//
//   plain -url http://example.com -file example-output.txt
//
func main() {
	url := flag.String("url", DefaultURL, "URL of the page you'd like to read")
	filepath := flag.String("file", "", "Optional filepath to output the page text")
	flag.Parse()

	text := makePlain(*url)

	if *filepath != "" {
		if err := ioutil.WriteFile(*filepath, []byte(text), 0666); err != nil {
			fmt.Printf("Failed to write text to '%s'\n", *filepath)
			log.Error(err)
		} else {
			fmt.Printf("Text successfully written to '%s'\n", *filepath)
		}
	} else {
		fmt.Println(text)
	}
}

// Given a URL, extracts the text we care about and returns it as a string ("make it plain!")
func makePlain(url string) string {
	response, err := loadPage(url)
	if err != nil {
		log.Error(err)
	}

	text, err := extractText(response)
	if err != nil {
		log.Error(err)
	}

	return text
}

// Sends an HTTP request to the specified URL and returns the response.
func loadPage(url string) (*http.Response, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %s", response.Status)
	}

	return response, nil
}

// Given an HTTP response, finds all HTML "text" tags and extracts their text content.
// What we consider a "text tag" is defined in the `TextSelector` constant. Returns a
// plaintext string of all the extracted text.
func extractText(response *http.Response) (string, error) {
	if response == nil {
		return "", errors.New("Nothing to see here!")
	}

	defer response.Body.Close()
	var textContents []string

	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return "", err
	}

	doc.Find(TextSelector).Each(func(i int, s *goquery.Selection) {
		textContents = append(textContents, formatText(s))
	})

	return strings.Join(textContents, "\n\n"), nil
}

// Extracts and formats the text from a selected HTML tag. We capitalize headers, and
// remove extra newlines that may be in paragraphs.
func formatText(s *goquery.Selection) string {
	if s == nil {
		return ""
	}

	var text string

	switch s.Nodes[0].Data {
	case "p":
		text = strings.ReplaceAll(s.Text(), "\n", " ")
	case "h1", "h2", "h3", "h4", "h5", "h6":
		text = strings.ToUpper(s.Text())
	}

	return text
}
