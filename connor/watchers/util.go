package watchers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func fetchBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	//I'm not sure about this function. Don't bite.
	if resp.StatusCode != http.StatusOK {
		log.Printf("request to %s returns status %d: %s", url, resp.StatusCode, err)

		for i := 0; i < 5; i++ {
			log.Printf("try to connect nanopool API %v", i)

			status, err := retryConnection(url)
			if err != nil {
				return nil, err
			}

			if status == true {
				log.Printf("nanopool return %v", http.StatusOK)
				return ioutil.ReadAll(resp.Body)
			}
		}
		return nil, fmt.Errorf("connection limit expired. request to %s returns status %d: %s", url, resp.StatusCode, err)
	}
	return ioutil.ReadAll(resp.Body)
}

func retryConnection(url string) (bool, error) {
	time.Sleep(2 * time.Minute)

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, nil
	}
}
