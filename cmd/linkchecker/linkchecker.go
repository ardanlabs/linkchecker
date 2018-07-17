package main

import (
	"bufio"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

type CheckResult struct {
	HTTPCode int
	Referrer string
	Error    error
	Body     string
	Recursed bool
}

var (
	linksChecked map[string]*CheckResult
	host         string
	timeout      int

	// Compiled regular expressions to use.
	reImage       *regexp.Regexp
	reURL         *regexp.Regexp
	reCurrentHost *regexp.Regexp
	reURLAbsolute *regexp.Regexp
	reURLRelative *regexp.Regexp
)

func main() {
	flag.StringVar(&host, "host", "", "Hostname and port of site to check.")
	flag.IntVar(&timeout, "timeout", 5, "Timeout in seconds.")
	flag.Parse()

	// Just add http:// to the host name and go.
	link := "http://" + host
	log.Println("Checking:", link)

	// Map that will hold all the link results.
	linksChecked = make(map[string]*CheckResult)

	// Compile regular expressions to be used.
	reCurrentHost = regexp.MustCompile("http(s)?://(www\\.)?" + host + ".*")
	reURL = regexp.MustCompile("http(s)?://.*")
	reImage = regexp.MustCompile("(jpg|svg|gif|png|js)(\\?.*)?$")
	reURLAbsolute = regexp.MustCompile("(src|href)=('|\")(?P<url>http(s)?://[^\"']*)('|\")")
	reURLRelative = regexp.MustCompile("(src|href)=('|\")(?P<url>/[^\"']*)('|\")")

	// If .linkignore file exists add links to checked result.
	if _, err := os.Stat(".linkignore"); err == nil {
		func(links map[string]*CheckResult) {
			file, err := os.Open(".linkignore")
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			// Read each line from file into a check result.
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()

				cr := &CheckResult{
					Referrer: "Ignored",
					HTTPCode: 100,
				}

				// If URL doesn't include http(s) then add domain.
				if !reURL.Match([]byte(line)) {
					line = link + line
				}

				links[line] = cr
			}
		}(linksChecked)
	}

	// Download the root page.
	start := time.Now()
	cr := download("", link)
	if cr.Error != nil {
		log.Fatal(cr.Error)
	}
	linksChecked[link] = cr

	// Recurse through the rest of the site.
	recurse(link, cr.Body)

	// Summarize results.
	log.Println("--------------------------------------------------------------")
	log.Println("These links where ignored.")
	for link, cr := range linksChecked {
		if cr.HTTPCode == 100 {
			// Log the errors again at the bottom for convience.
			var errStr string
			if cr.Error != nil {
				errStr = cr.Error.Error()
			}
			log.Printf("Referrer: %s Link: %s HTTPCode: %d %s\n", cr.Referrer, link, cr.HTTPCode, errStr)
		}
	}
	log.Println("--------------------------------------------------------------")
	log.Println("These links didn't check out.")
	var fives, fours, threes, twos, ones, errors int
	for link, cr := range linksChecked {
		switch {
		case cr.HTTPCode >= 500:
			fives++
		case cr.HTTPCode >= 400:
			fours++
		case cr.HTTPCode >= 300:
			threes++
		case cr.HTTPCode >= 200:
			twos++
		case cr.HTTPCode >= 100:
			ones++
		default:
			errors++
		}

		if cr.HTTPCode > 299 {
			// Log the errors again at the bottom for convience.
			var errStr string
			if cr.Error != nil {
				errStr = cr.Error.Error()
			}
			log.Printf("Referrer: %s Link: %s HTTPCode: %d %s\n", cr.Referrer, link, cr.HTTPCode, errStr)
		}
	}

	dur := time.Since(start)
	log.Println("--------------------------------------------------------------")
	log.Printf("Duration: %.0fs", dur.Seconds())
	log.Printf("Results 500s: %d 400s: %d 300s: %d 200s: %d Errors: %d Ignored: %d",
		fives, fours, threes, twos, errors, ones)

	if fives+fours+threes+errors > 0 {
		os.Exit(1)
	}
}

// recurse parses the html passed for urls, it takes the referrer link
// to build relative links.
func recurse(link, html string) {
	// Parse all the links from the html
	ls := parseLinks(link, html)

	// Loop through all the links and download asynchronously.
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}
	for _, l := range ls {

		// If link already checked continue.
		if _, ok := linksChecked[l]; ok {
			continue
		}

		// Download in a new routine.
		wg.Add(1)
		go func(referrer, link string) {
			defer wg.Done()
			cr := download(referrer, link)

			// Write result to links checked map.
			mutex.Lock()
			linksChecked[link] = cr
			mutex.Unlock()

			log.Printf("Referrer: %s Link: %s HTTPCode: %d\n", cr.Referrer, link, cr.HTTPCode)
		}(link, l)
	}
	wg.Wait()

	linksChecked[link].Recursed = true

	// Loop through the downloaded links and recurse
	for _, l := range ls {
		// If image don't recurse, continue to next link.
		if !isHTML(l) {
			continue
		}

		// If the link has not been recursed yet and for current host
		// then recurse through it.
		if !linksChecked[l].Recursed {
			if reCurrentHost.Match([]byte(l)) {
				recurse(l, linksChecked[l].Body)
			}
		}
	}
}

// isHTML returns true if a url is for an image.
func isHTML(url string) bool {
	if reImage.Match([]byte(url)) {
		return false
	}
	return true
}

// download gets the url passed returns an error or the html
// and the status code.
func download(referrer, url string) *CheckResult {
	cr := &CheckResult{Referrer: referrer}

	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	// If image or js don't download body.
	if !isHTML(url) {
		response, err := client.Head(url)
		if err != nil {
			cr.Error = errors.New("Error getting header: " + err.Error())
			return cr
		}

		cr.HTTPCode = response.StatusCode
		return cr
	}

	// Download HTML.
	retries := 3
	for ; retries > 0; retries-- {
		response, err := client.Get(url)
		if err != nil {
			cr.Error = errors.New("Error getting header: " + err.Error())
			continue
		}
		cr.HTTPCode = response.StatusCode

		// Download HTML body.
		defer response.Body.Close()
		b, err := ioutil.ReadAll(response.Body)
		if err != nil {
			cr.Error = errors.New("Error downloading: " + err.Error())
			continue
		}
		cr.Body = string(b)
	}

	return cr
}

// parseLinks parses html s for urls and returns them as a slice.
func parseLinks(link, s string) []string {
	u, err := url.Parse(link)
	if err != nil {
		log.Println(err, ":", link)
	}
	var links []string

	// Get anything that looks like an absolute url.
	for _, l := range reURLAbsolute.FindAllSubmatch([]byte(s), -1) {
		links = append(links, string(l[3]))
	}

	// Get anything that looks like a relative url.
	// Add the hostname.
	for _, l := range reURLRelative.FindAllSubmatch([]byte(s), -1) {
		nl := string(l[3])

		// If starts with // then use the same scheme but not really
		// a relative link.
		if len(nl) > 1 && string(nl[0:2]) == "//" {
			links = append(links, u.Scheme+":"+nl)
			continue
		}

		// Relative link use the same scheme and host.
		links = append(links, u.Scheme+"://"+u.Host+nl)
	}

	return links
}
