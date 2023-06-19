package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var idRegExp = regexp.MustCompile(`/.*?/(.*?).m3u8`)
var hlsRegExp = regexp.MustCompile(`"hlsManifestUrl":"(.*?)"`)

func getLiveStream(url string) (string, error) {
	fmt.Println("youtube url:", url)
	var res, err = http.Get(url)

	if err != nil {
		return "", err
	}

	fmt.Println("response status code", res.StatusCode)

	if res.StatusCode == http.StatusOK {
		var body, err = io.ReadAll(res.Body)

		if err != nil {
			return "", err
		}

		res.Body.Close()

		var re = regexp.MustCompile(`"hlsManifestUrl":"(.*?)"`)
		var matches = re.FindStringSubmatch(string(body))

		if len(matches) > 1 {
			var stream = matches[1]
			if stream == "" {
				return "", errors.New("stream url cannot be empty")
			} else {
				fmt.Println("live stream:", stream)
				return stream, nil
			}
		} else {
			return "", errors.New("stream url not found on youtube page: " + url)
		}
	} else {
		return "", errors.New("error response code from youtube: " + strconv.Itoa(res.StatusCode))
	}

}

func find(text string, re regexp.Regexp) (string, error) {
	var matches = re.FindStringSubmatch(text)

	if len(matches) > 1 {
		var match = matches[1]
		if match == "" {
			return "", errors.New("match cannot be empty")
		} else {
			fmt.Println("match:", match)
			return match, nil
		}
	} else {
		return "", errors.New("match not found")
	}
}

func getId(path string) (string, error) {
	var id, err = find(path, *idRegExp)

	if err != nil {
		return "", err
	}

	return id, nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	fmt.Println("request url: ", path)

	if strings.HasPrefix(path, "/channel/") {
		var id, err = getId(path)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			fmt.Println("request for channel with id: ", id)

			var stream, err = getLiveStream("https://www.youtube.com/channel/" + id + "/live")

			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Location", stream)
				w.WriteHeader(http.StatusFound)
			}
		}
	} else if strings.HasPrefix(r.URL.Path, "/video/") {
		var id, err = getId(path)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			fmt.Println("request for video with id: ", id)

			var stream, err = getLiveStream("https://www.youtube.com/watch?v=" + id)

			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Header().Set("Location", stream)
				w.WriteHeader(http.StatusFound)
			}
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

func main() {
	http.HandleFunc("/", handleRequest)
	fmt.Println("http server starting on port 3333")
	var err = http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}