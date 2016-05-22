package yelp

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/cathalgarvey/sqrape"
	"github.com/hashicorp/golang-lru"
)

var cache *lru.ARCCache

func init() {
	cache, _ = lru.NewARC(128)
}

func getPage(url string) (r *http.Response, err error) {
	return http.Get(url)
}

func isPaginate(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	start := u.Query().Get("start")
	return start != "" && start != "0"
}

func NewBusiness(url string) (b LocalBusiness, err error) {
	b.URL = url
	r, err := getPage(url)
	if err != nil {
		return b, err
	}
	err = sqrape.ExtractHTMLReader(r.Body, &b, isPaginate(url))
	return b, err
}

// FetchReviews aggregates all reviews for the business.
func (b *LocalBusiness) FetchReviews() {
	if cache.Contains(b.URL) {
		log.Printf("found business reviews for %s in cache\n", b.Name)
		val, _ := cache.Get(b.URL)
		b.Reviews = val.([]Review)
		return
	}

	wg := &sync.WaitGroup{}
	log.Printf("fetching business reviews %s\n", b.Name)

	for url := range b.paginationURLs() {
		wg.Add(1)
		go func(url string) {
			log.Printf("fetching reviews on url %s\n", url)
			defer wg.Done()

			p, err := NewBusiness(url)
			if err != nil {
				return
			}

			b.Reviews = append(b.Reviews, p.Reviews...)
			log.Printf("done fetching reviews on url %s\n", url)
		}(url)
	}

	wg.Wait()
	cache.Add(b.URL, b.Reviews)
	log.Printf("added business reviews for %s to cache\n", b.Name)
}

// FilterReviews filters down the list of reviews based on the provided filters.
func (b *LocalBusiness) FilterReviews(filters []ReviewFilter) (err error) {
	filterFuncs := make([]reviewFilterFunc, 0, len(filters))
	for _, f := range filters {
		switch f.Type {
		case "min_review_length":
			length, _ := strconv.Atoi(f.Value)
			filterFuncs = append(filterFuncs, makeFilterMinReviewLength(length))

		case "min_author_reviews":
			length, _ := strconv.Atoi(f.Value)
			filterFuncs = append(filterFuncs, makeFilterMinAuthorReviews(length))

		case "max_proximity":
			filterFuncs = append(filterFuncs, makeFilterMaxProximity(f.Value, b.Address))

		default:
			log.Printf("filter %v unsupported\n", f)
			continue
		}
		log.Printf("filter %v generated\n", f)
	}

	// TODO: Why can't we just use the existing slice?
	filteredReviews := make([]Review, 0, len(b.Reviews))
	for _, r := range b.Reviews {
		if r.matchAny(filterFuncs) {
			// log.Printf("review #%d matched filter, skipping: %v\n", i, r)
			continue
		}
		filteredReviews = append(filteredReviews, r)
	}

	b.Reviews = filteredReviews
	return
}

type reviewFilterFunc func(r *Review) bool

// matchAny returns a boolean if the review matches any of the provided filters.
func (r *Review) matchAny(filters []reviewFilterFunc) bool {
	for _, filter := range filters {
		if filter(r) == true {
			return true
		}
	}
	return false
}

func makeFilterMaxProximity(level string, addr Address) reviewFilterFunc {
	var suffix string

	switch level {
	case "LOCALITY":
		suffix = fmt.Sprintf("%s, %s", addr.Locality, addr.Region)
	case "REGION":
		suffix = fmt.Sprintf(", %s", addr.Region)
	}

	return func(r *Review) bool {
		return !strings.HasSuffix(r.Author.Location, suffix)
	}
}

func makeFilterMinReviewLength(n int) reviewFilterFunc {
	return func(r *Review) bool {
		return len(r.Description) < n
	}
}

func makeFilterMinAuthorReviews(n int) reviewFilterFunc {
	return func(r *Review) bool {
		return r.Author.ReviewCount < n
	}
}

// CalculateRating returns the newly calculated rating score.
//
// It only takes into consideration the number of reviews it has in-memory.
func (b *LocalBusiness) CalculateRating() float64 {
	if len(b.Reviews) == 0 {
		return 0.0
	}

	var sum float64
	for _, r := range b.Reviews {
		sum += r.Rating
	}
	return sum / float64(len(b.Reviews))
}

func (b LocalBusiness) paginationURLs() (urls chan string) {
	urls = make(chan string, 64)
	go func() {
		for i := 0; i < b.ReviewCount; i += 20 {
			urls <- fmt.Sprintf("%s?start=%d", b.URL, i)
		}
		close(urls)
	}()
	return urls
}
