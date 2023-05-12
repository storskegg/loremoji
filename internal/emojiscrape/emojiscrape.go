package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

const unicodeEmojiURL = "https://www.unicode.org/emoji/charts/full-emoji-list.html"
const codeSelector = "td.code"

var exitCode = 0

type collector struct {
	codes map[int]struct{}
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

	cc := NewCodesCollector()

	c := colly.NewCollector(colly.AllowURLRevisit())

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
			fmt.Printf("[%d] %d\n", count, int(i))
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
}

/**
 * The Javascript query to implement in the scraper:
 *
 * [...document.querySelectorAll('.code')].map((v, k) => parseInt(v.innerText.split(' ')[0].replace('U+', '0x'), 16)).filter(a => a > 100000).sort()
 */
