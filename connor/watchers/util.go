package watchers

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func fetchBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s returns status %d", url, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
