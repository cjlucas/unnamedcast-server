package itunes

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func allSelections(s *goquery.Selection) []*goquery.Selection {
	var out []*goquery.Selection
	s.Each(func(i int, s *goquery.Selection) {
		out = append(out, s)
	})
	return out
}

var alphabetList = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ#", "")
var itunesRssFeedRegexp = regexp.MustCompile(`"feedUrl":"(https?:\/\/[^"]*)`)
var httpClient = http.Client{}

func AlphabetPageListForFeedListPage(urlStr string) ([]string, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var out []string
	vals := url.Query()

	for i := range alphabetList {
		vals.Set("letter", alphabetList[i])
		url.RawQuery = vals.Encode()
		out = append(out, url.String())
	}

	return out, nil
}

func itunesHTTPGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Connection", "Close")
	req.Header.Set("User-Agent", "iTunes/12.3.2.0")

	return req, nil
}

func docFromHTTPReq(req *http.Request) (*goquery.Document, error) {
	c := http.Client{}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Received non-200 status code %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func docFromURL(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return docFromHTTPReq(req)
}

type goqueryHelper struct {
	d *goquery.Document
}

func (p *goqueryHelper) findSelections(selector string) []*goquery.Selection {
	return allSelections(p.d.Find(selector))
}

type FeedListPage struct {
	goqueryHelper
}

func NewFeedListPage(url string) (*FeedListPage, error) {
	doc, err := docFromURL(url)
	if err != nil {
		return nil, err
	}

	return &FeedListPage{
		goqueryHelper{
			d: doc,
		},
	}, nil
}

func (p *FeedListPage) PaginationPageList() []string {
	var out []string

	for _, s := range p.findSelections(".paginate li a") {
		if href, ok := s.Attr("href"); ok {
			out = append(out, href)
		}
	}

	return out
}

func (p *FeedListPage) FeedURLs() []string {
	var urls []string

	for _, s := range p.findSelections("#selectedcontent a") {
		if href, ok := s.Attr("href"); ok {
			urls = append(urls, href)
		}
	}

	return urls
}

type GenreListPage struct {
	goqueryHelper
}

func NewGenreListPage() (*GenreListPage, error) {
	doc, err := docFromURL("https://itunes.apple.com/us/genre/podcasts/id26?mt=2")
	if err != nil {
		return nil, err
	}

	return &GenreListPage{
		goqueryHelper{
			d: doc,
		},
	}, nil
}

func (p *GenreListPage) GenreURLs() []string {
	var out []string
	for _, s := range p.findSelections(".top-level-genre") {
		if href, ok := s.Attr("href"); ok {
			out = append(out, href)
		}
	}
	return out
}

func ResolveiTunesFeedURL(url string) (string, error) {
	req, err := itunesHTTPGetRequest(url)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		return "", err
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("Received unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if matches := itunesRssFeedRegexp.FindSubmatch(data); len(matches) > 1 {
		return string(matches[1]), nil
	}

	return "", errors.New("No match for feed url found in response body")
}
