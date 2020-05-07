package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func run() error {
	yearStartPtr := flag.Int("yearstart", 2018, "the year for popular names")
	yearEndPtr := flag.Int("yearend", 2018, "the year for popular names")
	topPtr := flag.Int("top", 1000, "the number of names to fetch")
	sexPtr := flag.String("sex", "", "\"boy\" or \"girl\", required")

	flag.Parse()

	var tableMod int
	switch *sexPtr {
	case "boy":
		tableMod = 1
	case "girl":
		tableMod = 2
	case "":
		return errors.New("-sex is a required flag")
	default:
		return fmt.Errorf("Unknown value for -sex found (must be \"boy\" or \"girl\")")
	}

	if *topPtr != 20 && *topPtr != 50 && *topPtr != 100 && *topPtr != 500 && *topPtr != 1000 {
		return errors.New("top must be 20, 50, 100, 500, or 1000")
	}

	names := map[string]struct{}{}
	for year := *yearStartPtr; year <= *yearEndPtr; year++ {
		body := url.Values{}
		body.Add("year", fmt.Sprintf("%d", year))
		body.Add("top", fmt.Sprintf("%d", *topPtr))

		resp, err := (&http.Client{}).Post("https://www.ssa.gov/cgi-bin/popularnames.cgi", "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Found error code %s", resp.Status)
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return err
		}

		doc.Find("table").Each(func(i int, node *goquery.Selection) {
			if val, exists := node.Attr("summary"); !exists || val != fmt.Sprintf("Popularity for top %d", *topPtr) {
				return
			}

			node.Find("td").Each(func(i int, node *goquery.Selection) {
				if i%3 != tableMod {
					return
				}
				names[node.Text()] = struct{}{}
			})
		})
	}
	if len(names) == 0 {
		return errors.New("no names found")
	}

	nameList := make([]string, len(names))
	i := 0
	for name, _ := range names {
		nameList[i] = name
		i++
	}

	sort.Strings(nameList)
	for _, name := range nameList {
		fmt.Println(name)
	}
	return nil
}
