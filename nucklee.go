package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type CurlData struct {
	url          string
	method       string
	responseBody string
}

func main() {
	command, payload := argParser()
	dispatch(command, payload)
}

// argParser gets arguments passed to commnd line and returns two strings
// the first string is the <command> and the second string is the <payload>
// the grammar is nucklee <command> <payload> where <payload> is optional
func argParser() (string, string) {
	args := os.Args
	if len(args) >= 3 {
		return args[1], args[2]
	}
	if len(args) == 2 {
		return args[1], ""
	}
	if len(args) == 1 {
		return "", ""
	}
	return "", ""
}

//dispatch parses arguments to invoke the correct functionality
func dispatch(command, payload string) {
	command = strings.ToLower(command)
	if command == "start" {
		if payload == "" {
			log.Println("No curl repository specified, using current dir")
			payload = "."
		}
		start(payload)
	} else if command == "stop" {
		stop()
	} else if command == "" {
		nothing()
	} else if command != "" {
		unknown()
	} else {
		panic("Unknown condition")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi love %s", r.URL.Path[1:])
}

func start(dir string) {
	cfs := filterCurlFiles(dir)
	cache := buildCache(cfs)

	if len(cache) > 0 {
		log.Println("cache", len(cache))
		http.HandleFunc("/", handler)
		log.Fatal(http.ListenAndServe(":12345", nil))
	}
}

func stop() {
	log.Println("Stop")
}

func nothing() {
	log.Println("Nothing to do")
}

func unknown() {
	log.Println("Unknown command")
}

func filterCurlFiles(dir string) []string {
	curlFiles := make([]string, 0)

	isCurlFile := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".curl") {
			fp, _ := filepath.Abs(filepath.Dir(path) + "/" + fi.Name())
			curlFiles = append(curlFiles, fp)
		}
		return nil
	}

	err := filepath.Walk(dir, isCurlFile)
	check(err)
	return curlFiles
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readFile(fpath string) string {
	bytes, err := ioutil.ReadFile(fpath)
	check(err)
	return string(bytes)
}

func extractHTTPURL(curlLine string) (string, error) {
	searchPhrase := "Rebuilt URL to:"
	i := strings.Index(curlLine, searchPhrase)
	if i > -1 {
		i = i + len(searchPhrase)
		return strings.TrimSpace(curlLine[i:]), nil
	} else {
		return "", errors.New("No URL found")
	}
}

func findResponseStartLine(lines []string) (int, error) {
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 1 && l == "<" {
			return i, nil
		}
	}
	return -1, errors.New("No response found")
}

func findHTTPMethod(lines []string) (string, error) {
	for _, l := range lines {
		if strings.HasPrefix(l, ">") {
			words := strings.Split(l, " ")
			return words[1], nil
		}
	}
	return "", errors.New("HTTP Method not found")
}

func hasBodySize(line string) bool {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "]") {
		return true
	}
	return false
}

func parseFile(data string) (*CurlData, error) {
	var cd CurlData
	lines := strings.Split(data, "\n")
	url, err := extractHTTPURL(lines[0])
	if err == nil {
		log.Println(url)
		cd := new(CurlData)
		cd.url = url
		httpMethod, err := findHTTPMethod(lines)
		cd.method = httpMethod
		if err == nil {
			responseStartLine, err := findResponseStartLine(lines)
			if err == nil {
				response := lines[responseStartLine+1:]
				if hasBodySize(response[0]) {
					response = response[1:]
					cd.responseBody = strings.Join(response, "\n")
				}
				return cd, nil
			}
		}
	}
	return &cd, err
}

func buildCache(files []string) map[string]*CurlData {
	cache := make(map[string]*CurlData)
	for _, fp := range files {
		s := readFile(fp)
		cd, err := parseFile(s)
		if err == nil {
			log.Println(cd.url, len(cd.responseBody))
			cache[cd.url] = cd
		}
	}
	return cache
}
