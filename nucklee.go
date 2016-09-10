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

type requestData struct {
	url    string
	method string
}

type responseData struct {
	headers map[string]string
	body    string
}

var cache = make(map[requestData]responseData)
var lineBreak = "##"

// Main function only for testing. This should be testing Load()
func main() {
	projectPath := argParser()
	Load(projectPath)

	if len(cache) == 0 {
		return
	}

	fmt.Println(cache)

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":12345", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	var request = new(requestData)
	request.url = r.URL.Path
	request.method = r.Method
	fmt.Fprintf(w, "URL: %s, Response: %s", r.URL.Path, cache[*request])
}

func argParser() string {
	args := os.Args
	defaultpath := "/usr/local/nucklee"
	if len(args) >= 2 {
		return args[1]
	}

	return defaultpath
}

// Load function loads all the valid http files from the project path and
// stores the serialized object.
func Load(projectPath string) error {
	err := filepath.Walk(projectPath, processPath)
	return err
}

func processPath(path string, pathInfo os.FileInfo, _ error) error {
	if !pathInfo.IsDir() && isHTTPFile(pathInfo.Name()) {
		cacheRequests(path)
	}
	return nil
}

func isHTTPFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".http")
}

func getHTTPMethod(fileName string) (string, error) {
	if fileName[:len(fileName)-5] == "" {
		return "", errors.New("Invalid File Name")
	}

	return fileName[:len(fileName)-5], nil
}

func cacheRequests(path string) error {
	contents := readFile(path)
	err := parseFile(contents)

	return err
}

func readFile(fpath string) string {
	bytes, err := ioutil.ReadFile(fpath)
	check(err)
	return string(bytes)
}

func parseFile(data string) error {
	items := strings.Split(data, lineBreak)
	for i := 0; i < len(items); i++ {
		err := processItem(items[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func processItem(item string) error {
	lines := strings.Split(strings.TrimSpace(item), "\n")
	request, err := extractHTTPRequestData(lines[0])
	if err == nil {
		responseStartLine, err := findResponseStartLine(lines)
		if err == nil {
			response, err := getResponse(lines[responseStartLine:])
			if err == nil {
				cache[*request] = *response
				return nil
			}
		}
	}
	return err
}

func extractHTTPRequestData(requestLine string) (*requestData, error) {
	request := new(requestData)
	requestPieces := strings.Split(requestLine, " ")
	if len(requestPieces) == 3 {
		request.method = requestPieces[0]
		request.url = requestPieces[1]

		return request, nil
	}

	return request, errors.New("No URL found")
}

func findResponseStartLine(lines []string) (int, error) {
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "HTTP") {
			return i, nil
		}
	}
	return -1, errors.New("No response found")
}

func getResponse(responseLines []string) (*responseData, error) {
	headers := make(map[string]string)
	var i int
	response := new(responseData)

	for i = 2; i < len(responseLines); i++ {
		responseLines[i] = strings.TrimSpace(responseLines[i])
		if responseLines[i] == "" {
			break
		}

		header := strings.Split(responseLines[i], ": ")
		if len(header) != 2 {
			return response, errors.New("Invalid herader")
		}

		headers[header[0]] = header[1]
	}

	response.headers = headers
	response.body = strings.Join(responseLines[i:], "\n")

	return response, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
