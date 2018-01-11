package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const rssFeed = "https://www.nzz.ch/startseite.rss"

func handler(w http.ResponseWriter, r *http.Request) {
	text, err := fetch()
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", text)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func fetch() (string, error) {
	start := time.Now()
	resp, err := http.Get(rssFeed)
	if err != nil {
		return "", fmt.Errorf("couldn't fetch feed: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't read response: %v", err)
	}
	items := regexp.MustCompile(`<item>.*?</item>`)
	original := string(body)
	findings := items.FindAllString(original, -1)
	var res bytes.Buffer
	skipped := 0
outer:
	for _, item := range findings {
		for _, domain := range blacklist {
			if strings.Contains(item, "nzz.ch/"+domain) {
				skipped++
				continue outer
			}
		}
		res.WriteString(item)
	}
	allItems := regexp.MustCompile(`<item>.*</item>`)
	modified := allItems.ReplaceAllString(original, res.String())
	log.Printf("request took %v (%d of %d items skipped)", time.Since(start), skipped, len(findings))
	return modified, nil
}

var blacklist = []string{
	"sport",
	"briefing",
	"feuilleton",
	"panorama",
	"zuerich",
	"mobilitaet",
	"gesellschaft",
	"digital",
	"meinung",
	"leserdebatte",
	"video",
}
