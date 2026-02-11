package infrastructure

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TrendPoint struct {
	Timestamp string `json:"timestamp"`
	Value     int    `json:"value"`
}

type Result struct {
	SpamTrap    int          `json:"spam_trap"`
	Blocklists  string       `json:"blocklists"`
	Complaints  string       `json:"complaints"`
	SenderScore int          `json:"sender_score"`
	SSTrend     []TrendPoint `json:"ss_trend"`
	SSVolume    []TrendPoint `json:"ss_volume"`
}

type Parser struct {
	source string
}

func NewParser(source string) *Parser {
	return &Parser{
		source: source,
	}
}

func (p *Parser) Parse() *Result {
	reader := strings.NewReader(p.source)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	result := Result{}

	doc.Find("#repTable tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() == 2 {
			key := strings.TrimSpace(cells.Eq(0).Text())
			valStr := strings.TrimSpace(cells.Eq(1).Text())

			switch key {
			case "Spam Traps":
				val, _ := strconv.Atoi(valStr)
				result.SpamTrap = val
			case "Blocklists":
				result.Blocklists = valStr
			case "Complaints":
				result.Complaints = valStr
			}
		}
	})
	scriptText := ""
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "ssData.ss_trend") {
			scriptText = s.Text()
		}
	})

	reScore := regexp.MustCompile(`ssData.senderscore = ([0-9]+);`)
	scoreMatches := reScore.FindStringSubmatch(scriptText)
	if len(scoreMatches) > 1 {
		result.SenderScore, _ = strconv.Atoi(scoreMatches[1])
	}

	reTrend := regexp.MustCompile(`ssData.ss_trend = \[([^\]]+)\];`)
	trendMatches := reTrend.FindStringSubmatch(scriptText)

	if len(trendMatches) > 1 {
		jsonStr := "[" + trendMatches[1] + "]"
		err := json.Unmarshal([]byte(jsonStr), &result.SSTrend)
		if err != nil {
			fmt.Println("Error parse JSON trends:", err)
		}
	}

	reVolume := regexp.MustCompile(`ssData.ss_volume_trend = \[([^\]]+)\];`)
	volumeMatches := reVolume.FindStringSubmatch(scriptText)

	if len(volumeMatches) > 1 {
		jsonStr := "[" + volumeMatches[1] + "]"
		err := json.Unmarshal([]byte(jsonStr), &result.SSVolume)
		if err != nil {
			fmt.Println("Error parse JSON volume:", err)
		}
	}

	return &result
}
