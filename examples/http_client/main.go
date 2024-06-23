package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/initlevel5/workerpool"
	"go.uber.org/multierr"
)

const (
	numWorkers   = 10
	numUrlsParts = 10
)

type result struct {
	url     string
	sz      int
	err     error
	elapsed time.Duration
}

func fetchURL(client *http.Client, url string) (int, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return len(body), nil
}

func buildUrls() []string {
	urlsPart := []string{
		"http://ya.ru",
		"http://mail.ru",
		"http://rambler.ru",
		"http://rbc.ru",
		"http://google.ru",
		"http://lenta.ru",
		"http://go.dev",
		"http://stackoverflow.com",
		"http://dzen.ru",
		"http://wildberries.ru",
	}

	urls := make([]string, 0, numUrlsParts*len(urlsPart))
	for i := 0; i < numUrlsParts; i++ {
		urls = append(urls, urlsPart...)
	}

	return urls
}

func main() {
	var (
		err error
		wg  sync.WaitGroup
	)

	urls := buildUrls()

	pool := workerpool.New(numWorkers, len(urls))
	defer pool.Close()

	resultCh := make(chan *result, len(urls))

	client := &http.Client{}

	start := time.Now()

	for _, url := range urls {
		url := url

		wg.Add(1)
		pool.MustAddTask(func() {
			defer wg.Done()

			start := time.Now()

			sz, err := fetchURL(client, url)

			resultCh <- &result{url: url, sz: sz, err: err, elapsed: time.Since(start)}
		})
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		if res == nil {
			continue
		}
		if res.err != nil {
			err = multierr.Append(err, res.err)
			continue
		}
		fmt.Printf("%s %d %s\n", res.url, res.sz, res.elapsed)
	}

	if err != nil {
		fmt.Println("err: ", err)
	}

	fmt.Println("elapsed: ", time.Since(start))
}
