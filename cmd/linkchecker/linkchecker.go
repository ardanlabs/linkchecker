package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

type CheckResult struct {
	HTTPCode int
	Referrer string
	Error    error
}

var linksChecked map[string]CheckResult
var host string

func main() {
	flag.StringVar(&host, "host", "", "Hostname and port of site to check.")
	flag.Parse()

	// Just add http:// to the host name and go.
	link := "http://" + host
	log.Println("Checking:", link)

	// Map that will hold all the link results.
	linksChecked = make(map[string]CheckResult)

	// Download the root page.
	err, b, _ := download(link)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Recurse through the rest of the site.
	recurse(link, b)
}

// recurse parses the html passed for urls, it takes the referrer link
// to build relative links.
func recurse(link, html string) {
	// Parse all the links from the html
	ls := parseLinks(link, html)

	// Loop through all the links and download. Recurse again if html.
	for _, l := range ls {

		// If link not already checked, download.
		if _, ok := linksChecked[l]; !ok {
			err, b, status := download(l)
			cr := CheckResult{
				HTTPCode: status,
				Referrer: link,
				Error:    err,
			}
			linksChecked[l] = cr

			log.Printf("Referrer: %s Link: %s HTTPCode: %d\n", link, l, status)

			// If image don't recurse, continue to next link	.
			if isImage(l) {
				continue
			}

			// If link on host being checked then recurse through.
			r := fmt.Sprintf("http(s)?://(www\\.)?" + host + ".*")
			if found, _ := regexp.Match(r, []byte(l)); found {
				recurse(l, b)
			}
		}
	}
}

// isImage returns true if a url is for an image.
func isImage(url string) bool {
	if found, _ := regexp.Match("(jpg|svg|gif|png)$", []byte(url)); found {
		return true
	}
	return false
}

// download gets the url passed returns an error or the html
// and the status code.
func download(url string) (error, string, int) {
	response, err := http.Get(url)
	if err != nil {
		return err, "", 0
	}

	// If image don't download body.
	if isImage(url) {
		return nil, "", response.StatusCode
	}

	// Download html body.
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err, "", 0
	}

	return nil, string(b), response.StatusCode
}

// parseLinks parses html s for urls and returns them as a slice.
func parseLinks(link, s string) []string {
	u, err := url.Parse(link)
	if err != nil {
		log.Println(err, ":", link)
	}
	var links []string

	// Get anything that looks like an absolute url.
	r := regexp.MustCompile("('|\")http(s)?://[^\"']*\"")
	for _, l := range r.FindAllString(s, -1) {
		nl := l[1 : len(l)-1]
		links = append(links, nl)
	}

	// Get anything that looks like a relative url.
	// Add the hostname.
	r = regexp.MustCompile("\"/[^\"]*\"")
	for _, l := range r.FindAllString(s, -1) {
		nl := l[1 : len(l)-1]
		links = append(links, u.Scheme+"://"+u.Host+nl)
	}

	return links
}
