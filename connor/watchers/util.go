package watchers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const retryTimeout = 1 * time.Second

func fetchBody(url string) ([]byte, error) {
	body, err := fetchOnce(url)
	if err != nil {
		for i := 0; i < 5; i++ {
			time.Sleep(retryTimeout)

			body, err = fetchOnce(url)
			if err != nil {
				continue
			}

			return body, nil
		}

		return nil, fmt.Errorf("http connection retries exceeded, cannot perform http request to %s", url)
	}

	return body, nil
}

func fetchOnce(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for %s returns status %d", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
