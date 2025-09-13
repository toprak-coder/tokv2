package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

// TokenType defines the origin of the token within a URL.
type TokenType int

const (
	Subdomain TokenType = iota
	Path
	ParamName
	ParamValue
	Fragment
	Generic // For lines that cannot be parsed as URLs
)

// Token represents a word extracted from the input.
type Token struct {
	Value string
	Type  TokenType
}

var (
	minlength      int
	maxlength      int
	alphaNumOnly   bool
	filterString   string
	filterRegex    string
	pathOutputFile string
	paramOutputFile string

	regex *regexp.Regexp

	pathFile *os.File
	paramFile *os.File

	lastToken sync.Map
)

func main() {
	flag.IntVar(&minlength, "min", 1, "min length of string to be output")
	flag.IntVar(&maxlength, "max", 25, "max length of string to be output")
	flag.BoolVar(&alphaNumOnly, "alpha-num-only", false, "return only strings containing at least one letter and one number")
	flag.StringVar(&filterString, "f", "", "filter tokens to those containing this string")
	flag.StringVar(&filterRegex, "r", "", "filter tokens to those matching this regex pattern")
	flag.StringVar(&pathOutputFile, "o", "", "output file for path tokens")
	flag.StringVar(&paramOutputFile, "op", "", "output file for parameter name tokens")
	flag.Parse()

	var err error
	if filterRegex != "" {
		regex, err = regexp.Compile(filterRegex)
		if err != nil {
			log.Fatalf("invalid regex: %v", err)
		}
	}

	if pathOutputFile != "" {
		pathFile, err = os.Create(pathOutputFile)
		if err != nil {
			log.Fatalf("failed to create path output file: %v", err)
		}
		defer pathFile.Close()
	}

	if paramOutputFile != "" {
		paramFile, err = os.Create(paramOutputFile)
		if err != nil {
			log.Fatalf("failed to create param output file: %v", err)
		}
		defer paramFile.Close()
	}

	tokens := make(chan Token)
	var wg sync.WaitGroup

	// Worker goroutine to process tokens
	wg.Add(1)
	go func() {
		defer wg.Done()
		for token := range tokens {
			processToken(token)
		}
	}()

	// Main goroutine to read stdin and produce tokens
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		u, err := url.Parse(line)
		if err != nil || u.Scheme == "" || u.Host == "" {
			// Fallback to simple tokenization for non-URLs
			subTokenize(line, Generic, tokens)
			continue
		}

		// Hostname -> Subdomains
		hostname := u.Hostname()
		for _, part := range strings.Split(hostname, ".") {
			subTokenize(part, Subdomain, tokens)
		}

		// Path
		for _, part := range strings.Split(u.Path, "/") {
			subTokenize(part, Path, tokens)
		}

		// Query Parameters
		query, err := url.ParseQuery(u.RawQuery)
		if err == nil {
			for key, values := range query {
				subTokenize(key, ParamName, tokens)
				for _, value := range values {
					subTokenize(value, ParamValue, tokens)
				}
			}
		}
		
		// Fragment
		subTokenize(u.Fragment, Fragment, tokens)
	}

	close(tokens)
	wg.Wait()
}

// subTokenize performs basic tokenization on a string part.
func subTokenize(input string, tokenType TokenType, tokens chan<- Token) {
	var out strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			out.WriteRune(r)
		} else {
			if out.Len() > 0 {
				tokens <- Token{Value: out.String(), Type: tokenType}
				out.Reset()
			}
		}
	}
	if out.Len() > 0 {
		tokens <- Token{Value: out.String(), Type: tokenType}
	}
}

// processToken filters a token and writes it to the appropriate output.
func processToken(token Token) {
	str := token.Value
	
	// Deduplication
	if _, loaded := lastToken.LoadOrStore(str, true); loaded {
		return
	}

	// Filters
	if len(str) < minlength || len(str) > maxlength {
		return
	}

	if filterString != "" && !strings.Contains(str, filterString) {
		return
	}

	if regex != nil && !regex.MatchString(str) {
		return
	}

	if alphaNumOnly {
		hasLetter := false
		hasNumber := false
		for _, r := range str {
			if unicode.IsLetter(r) {
				hasLetter = true
			}
			if unicode.IsNumber(r) {
				hasNumber = true
			}
		}
		if !hasLetter || !hasNumber {
			return
		}
	}

	// Output dispatch
	switch token.Type {
	case Path:
		if pathFile != nil {
			fmt.Fprintln(pathFile, str)
			return
		}
	case ParamName:
		if paramFile != nil {
			fmt.Fprintln(paramFile, str)
			return
		}
	}

	// Default to stdout
	fmt.Println(str)
}
