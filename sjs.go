package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/janeczku/go-spinner"
)

// color print
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

func jsHttpPath(js string) string {
	find_start_http_path := regexp.MustCompile(`https?://[^/\s]+(?:/[^/\s]+)*`)
	matchesPathHTTP := find_start_http_path.FindStringSubmatch(js)
	return matchesPathHTTP[0]

}

// this func for find js main file
func extractJSFilesMainDomWeb(htmlContent string) []string {
	var jsFiless []string
	pattern := `"[^"|']*\/main[^"|']*\.js"`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(htmlContent, -1)
	if matches == nil {
		pattern := `("|')([^"']+\.js(?:\?[^"']*?)?)("|')`
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(htmlContent, -1)
		for _, match := range matches {
			if len(match) > 1 {

				mainJSRegex := regexp.MustCompile(`([^'|"]*\/main[^'|"]*\.js)`)
				matchesss := mainJSRegex.FindAllString(match[0], -1)
				if matchesss != nil {
					jsFiless = append(jsFiless, matchesss[0])
				}

			}
		}
	}

	return jsFiless
}

func domWebSite(url string) {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code: %d", response.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()
		jsdom := extractJSFilesMainDomWeb(scriptContent)
		if jsdom != nil {
			pathfileJSDom := jsdom[0]
			fmt.Println(pathfileJSDom)

		}

	})
}

func extractJSFilesMain(htmlContent string) []string {
	var jsFiles []string
	//(.*)main(.*)\.js$

	pattern := `([^"|']*main[^"|']*\.js)`
	regex := regexp.MustCompile(pattern)

	matches := regex.FindAllStringSubmatch(htmlContent, -1)
	if matches == nil {
		s := spinner.StartNew(Red + "[-] Can not find file main js [-]")
		time.Sleep(5 * time.Second)
		s.Stop()
		os.Exit(3)
	}
	// Extract matched URLs
	for _, match := range matches {
		if len(match) > 1 {
			jsFiles = append(jsFiles, match[0])
		} else {
			fmt.Println("this website not found main js try ...")
		}
	}

	return jsFiles
}
func removeDuplicates(s []string) []string {
	bucket := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := bucket[str]; !ok {
			bucket[str] = true
			result = append(result, str)
		}
	}
	return result
}
func requestsMain(url string) []string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	urls, err := client.Get(url)
	if err != nil {
		os.Exit(1)
	}
	body_url_read, _ := io.ReadAll(urls.Body)
	body := string(body_url_read)
	jsFiles := extractJSFilesMain(body)
	js := jsFiles[0]

	defer urls.Body.Close()

	findPathjs := regexp.MustCompile(`^(.+\.js)`)
	matchesPath := findPathjs.FindStringSubmatch(js)
	findname := strings.ReplaceAll(matchesPath[0], `src="`, "")
	mk_url := url + findname

	tlsoff := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// http requets api check status code
	clientjs := &http.Client{Transport: tlsoff}
	http_requets, err := clientjs.Get(mk_url)
	if err != nil {
		fmt.Println(err)
	}
	body_http, err := io.ReadAll(http_requets.Body)
	if err != nil {
		fmt.Println(err)
	}

	body_string := string(body_http)
	endpointRegex := regexp.MustCompile(`(?:\/)([a-zA-Z_][a-zA-Z0-9_/]+)(\/)[^"'><,;|(^/\\\[\])]*`)
	matchesJS := endpointRegex.FindAllStringSubmatch(body_string, -1)
	var craetelinkcheck []string
	for _, js := range matchesJS {
		input := js[0]
		clearInput := removeAllExtensions(input)
		removeback := removeAllback(clearInput)

		pattern := `.*.\..*`

		regexpPattern := regexp.MustCompile(pattern)

		result := regexpPattern.ReplaceAllString(removeback, "")
		replace_str_1 := strings.ReplaceAll(result, "`", "")
		replace_str_2 := strings.ReplaceAll(replace_str_1, `"`, "")
		craetelinkcheck = append(craetelinkcheck, replace_str_2)

	}
	result := removeDuplicates(craetelinkcheck)
	return result

}

func CheckEndpointStatusBurpProxy(proxy, urlapi, craetelinkcheck, headerinput, ms string) {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		os.Exit(1)
	}
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client_api := &http.Client{
		Transport: transport,
	}
	requestURL := urlapi + craetelinkcheck
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	if headerinput != "" {
		tokenGET := strings.Split(headerinput, ":")
		req.Header.Set(tokenGET[0], tokenGET[1])
	}

	resp, err := client_api.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	statudcode := resp.Status
	header := resp.Header
	stat := resp.StatusCode
	checkstatuscodefilter := strings.Split(ms, ",")
	if ms != "" {
		statacodesite := strconv.Itoa(stat)
		for _, int_code := range checkstatuscodefilter {
			if int_code == statacodesite {
				fmt.Printf(" => endpoint: %s\n => status code: %s\n", craetelinkcheck, statudcode)
				fmt.Println("Headers:")
				for key, val := range header {
					fmt.Printf("     %s: %s\n", key, val)
				}
				fmt.Println("\n\n\n")
			}
		}
	} else {
		fmt.Printf("  > endpoint: %s\n => status code: %s\n", craetelinkcheck, statudcode)
		fmt.Println("Headers:")
		for key, val := range header {
			fmt.Printf("     %s: %s\n", key, val)
		}
		fmt.Println("\n\n\n")
	}

}

func CheckEndpointStatus(urlapi, craetelinkcheck, headerinput, flag_ms string) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	requestURL := urlapi + craetelinkcheck
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	if headerinput != "" {
		tokenGET := strings.Split(headerinput, ":")
		req.Header.Set(tokenGET[0], tokenGET[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	statudcode := resp.Status
	header := resp.Header
	stat := resp.StatusCode
	checkstatuscodefilter := strings.Split(flag_ms, ",")

	statacodesite := strconv.Itoa(stat)
	if flag_ms != "" {
		for _, int_code := range checkstatuscodefilter {
			if int_code == statacodesite {
				fmt.Printf(" => endpoint: %s\n => status code: %s\n", craetelinkcheck, statudcode)
				fmt.Println("Headers:")
				for key, val := range header {
					fmt.Printf("     %s: %s\n", key, val)
				}
				fmt.Println("\n\n\n")
			}
		}
	} else {
		fmt.Printf(" => endpoint: %s\n => status code: %s\n", craetelinkcheck, statudcode)
		fmt.Println("Headers:")
		for key, val := range header {
			fmt.Printf("     %s: %s\n", key, val)
		}
		fmt.Println("\n\n\n")
	}

}

func removeAllback(input string) string {
	extensionPattern := `^" \./.*`
	re := regexp.MustCompile(extensionPattern)
	result := re.ReplaceAllString(input, "")
	return result
}
func removeAllExtensions(input string) string {
	extensionPattern := `\.[a-zA-Z0-9]+`
	re := regexp.MustCompile(extensionPattern)

	// Replace all occurrences of file extensions with an empty string
	result := re.ReplaceAllString(input, "")

	return result
}

func main() {
	name := flag.String("url", "", "url target: https://site.com")
	api := flag.String("api", "", "url api target: https://api.site.com")
	var ms string
	flag.StringVar(&ms, "ms", "", "this flag for filter status code ")
	var headerinput string
	flag.StringVar(&headerinput, "header", "", "this flag for set header requets")

	// flag.BoolVar(&useProxy, "proxy", false, "Use proxy")
	proxyURL := flag.String("proxy", "", "HTTP proxy address (e.g., http://127.0.0.1:8080)")

	// flag dom
	// dom := flag.Bool("dom", false, "flag dom defualt is false")
	flag.Parse()
	if *name == "" {
		fmt.Println(Red + "[-]Please set url for example: https://example.com ! Please used flag -help")
		os.Exit(1)
	}
	if *api == "" {
		fmt.Println(Red + "[-]Please set url api for example: https://api.example.com !")
		os.Exit(1)
	}

	s := spinner.StartNew(Green + "[+]EndPoint Scanner Started[+]")
	time.Sleep(5 * time.Second)
	s.Stop()

	craetelinkcheck := requestsMain(*name)

	for _, link := range craetelinkcheck {
		if *proxyURL != "" {

			CheckEndpointStatusBurpProxy(*proxyURL, *api, link, headerinput, ms)

		} else {
			CheckEndpointStatus(*api, link, headerinput, ms)

		}
	}
}
