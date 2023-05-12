package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

// maxBodySize is set to an unreasonably large number, because the file size as of this moment is 33.83M, and will
// likely only grow as more BS is added to the emoji codepages of the unicode standard. (What a waste.)
const maxBodySize = 128 * 1024 * 1024 // 128MB
const unicodeEmojiURL = "https://www.unicode.org/emoji/charts/full-emoji-list.html"
const codeSelector = "td.code"

var pkgName string
var out string

var exitCode = 0

type collector struct {
	codes map[int]struct{} // a map is used to dedupe
	mu    sync.Mutex
}

func NewCodesCollector() *collector {
	return &collector{
		codes: make(map[int]struct{}),
	}
}

func (c *collector) Push(code int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.codes[code] = struct{}{}
}

func (c *collector) Size() int {
	return len(c.codes)
}

func (c *collector) GetCodes() []int {
	c.mu.Lock()
	defer c.mu.Unlock()
	codes := make([]int, 0, len(c.codes))
	for code := range c.codes {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	return codes
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic Recovered:", r)
			os.Exit(100)
		}
		os.Exit(exitCode)
	}()

	flag.StringVar(&pkgName, "pkg", "", "target package name")
	flag.StringVar(&out, "out", "", "output file name")
	flag.Parse()

	cc := NewCodesCollector()

	c := colly.NewCollector(colly.AllowURLRevisit(), colly.MaxBodySize(maxBodySize))

	count := 0

	c.OnHTML(codeSelector, func(e *colly.HTMLElement) {
		s := strings.Split(e.Text, " ")[0]
		s = strings.ReplaceAll(s, "U+", "")

		i, err := strconv.ParseInt(s, 16, 32)
		if err != nil {
			fmt.Println(err)
			exitCode = 1
			return
		}

		if i > 100000 {
			count++
			cc.Push(int(i))
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	err := c.Visit(unicodeEmojiURL)
	if err != nil {
		fmt.Println(err)
		exitCode = 2
		return
	}

	c.Wait()

	fmt.Println("Iterated  : ", count)
	fmt.Println("Collected : ", cc.Size())

	sb := strings.Builder{}
	fmt.Fprintln(&sb, "")
}

/**
 * The Javascript query to implement in the scraper:
 *
 * [...document.querySelectorAll('.code')].map((v, k) => parseInt(v.innerText.split(' ')[0].replace('U+', '0x'), 16)).filter(a => a > 100000).sort()
 */

func validatePackageName(name string) bool {
	if name == "" {
		return false
	}

	if strings.ContainsAny(name, ".-") {
		return false
	}

	return true
}
