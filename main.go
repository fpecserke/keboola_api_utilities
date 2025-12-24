package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Request = struct {
	name   string
	method string
	url    string
}

type Token = struct {
	name string
	val  string
}

func main() {
	downloadMetadata := flag.Bool("m", false, "Download metadata")
	downloadTableDetails := flag.Bool("t", false, "Download table details")
	tokensFolder := flag.String("tokens-folder", "./tokens/", "Specify which folder to read tokens from")
	dataFolder := flag.String("data-folder", "./data/", "Specify which folder to write the downloaded data")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Downloads all kinds of data from Keboola based on your tokens...\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Each project has a different token but not all tokens are able to access everyhting inside a project\n")
		fmt.Fprintf(flag.CommandLine.Output(), "The name of the file wher the token is save will be used for foler structure of downloaded data\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// you have to download the tokens from Keboola and save them there
	// Token {name, val} will be populated like:
	// - name = File name
	// - val = content of the file
	tokens := ReadTokensFromDir(*tokensFolder)
	urls := []Request{
		{"00_Verify_Token", "GET", "https://connection.eu-central-1.keboola.com/v2/storage/tokens/verify"},
		{"01_List_Buckets", "GET", "https://connection.eu-central-1.keboola.com/v2/storage/buckets?include="},
		{"02_Worspaces", "GET", "https://connection.eu-central-1.keboola.com/v2/storage/workspaces"},
		{"03_Components", "GET", "https://connection.eu-central-1.keboola.com/v2/storage/components?include=configuration,rows,state"},
		{"04_Tables", "GET", "https://connection.eu-central-1.keboola.com/v2/storage/tables?include=metadata"},
	}

	if *downloadMetadata {
		var wg sync.WaitGroup
		wg.Add(len(tokens) * len(urls))
		// For each Keboola Project you need a new token, we iterate over them
		for _, token := range tokens {
			// write responses from each request to a file: ./data/{token.name}/{req.name}.json
			log.Printf("# %s:\n", token.name)
			for _, req := range urls {
				go func() {
					defer wg.Done()
					// log.Printf("- %s", req.name)
					response := SendRequestAndReturnBody(req, token.val)
					WriteToAndMakeFileIfNotExists(*dataFolder+token.name+"/"+req.name+".json", response)
				}()
			}
		}
		wg.Wait()
	}

	if *downloadTableDetails {
		for _, token := range tokens {
			tables := ReadTableInfoFromProjectFiles(token.name, *dataFolder)
			var wg2 sync.WaitGroup
			wg2.Add(len(tables))
			log.Printf("Downloading table info from project %q", token.name)
			// write table details from each table to a file: ./data/{token.name}/table_details/{table.ID}.json
			for _, table := range tables {
				go func() {
					defer wg2.Done()
					// log.Printf("- %s", table.ID)
					uniqueReq := Request{table.ID, "GET", table.URL}
					resp := SendRequestAndReturnBody(uniqueReq, token.val)
					WriteToAndMakeFileIfNotExists(*dataFolder+token.name+"/table_details/"+table.ID+".json", resp)
				}()
			}
			wg2.Wait()
		}
	}
}

func ReadTokensFromDir(path string) []Token {
	re, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Could not read directory %q | %v", path, err)
	}

	var tokens []Token
	for _, file := range re {
		name := file.Name()
		res, err := os.ReadFile(path + name)
		if err != nil {
			log.Fatalf("Could not read file '%s' | %v", name, err)
		}

		tokens = append(tokens, Token{name: name, val: string(res)})
	}

	return tokens
}

func SendRequestAndReturnBody(reqw Request, projectToken string) []byte {
	req, err := http.NewRequest(reqw.method, reqw.url, nil)
	if err != nil {
		log.Fatalf("sendRequest() | Couldnt create request | %v", err)
	}
	req.Header.Set("X-StorageApi-Token", strings.TrimSpace(projectToken))
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("sendRequest() | Could not get the url '%s' | %v", reqw.url, err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("sendRequest() | Couldnt read bytes | %v", err)
	}

	return bodyBytes
}

func WriteToAndMakeFileIfNotExists(fullPath string, content []byte) {
	err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(fullPath, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("Characters written: %d", len(content))
}

type Table = struct {
	URL string `json:"uri"`
	ID  string `json:"id"`
}

// ReadTableInfoFromProjectFiles ...
// this JSON parsing is done automatically by Go, because of how it's specified in the type Table
func ReadTableInfoFromProjectFiles(tokenName string, dataFolder string) []Table {
	fullpath := dataFolder + tokenName + "/04_Tables.json"
	re, err := os.ReadFile(fullpath)
	if err != nil {
		log.Fatalf("Could not read file '%s' | %v", fullpath, err)
	}
	var content []Table
	err = json.Unmarshal(re, &content)
	if err != nil {
		log.Fatalf("Could not parse file '%s' | %v", fullpath, err)
	}
	return content
}
