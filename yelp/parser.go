package yelp

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/namsral/microdata"
)

func getPage(url string) (r *http.Response, err error) {
	return http.Get(url)
}

// NewBusiness returns a new business parsed from the provided URL.
func NewBusiness(bizURL string) (b LocalBusiness, err error) {
	r, err := getPage(bizURL)
	if err != nil {
		return
	}
	defer r.Body.Close()

	u, err := url.Parse(bizURL)
	if err != nil {
		return
	}
	data, err := microdata.ParseHTML(r.Body, r.Header.Get("Content-Type"), u)

	b.URL = bizURL
	b.Name = strings.TrimSpace(data.Items[1].Properties["name"][0].(string))

	addr := data.Items[1].Properties["address"][0].(*microdata.Item).Properties
	b.Address = Address{
		StreetAddress: addr["streetAddress"][0].(string),
		Locality:      addr["addressLocality"][0].(string),
		Region:        addr["addressRegion"][0].(string),
		PostalCode:    addr["postalCode"][0].(string),
	}

	aggregateRating := data.Items[1].Properties["aggregateRating"][0].(*microdata.Item).Properties
	b.AggregateRating, err = strconv.ParseFloat(aggregateRating["ratingValue"][0].(string), 64)
	b.ReviewCount, err = strconv.Atoi(aggregateRating["reviewCount"][0].(string))
	b.Reviews = make([]Review, 0, b.ReviewCount)

	return
}

// FetchReviews aggregates all reviews for the business.
func (b *LocalBusiness) FetchReviews() {
	wg := &sync.WaitGroup{}
	log.Printf("fetching business reviews %s\n", b.Name)

	for url := range b.paginationURLs() {
		wg.Add(1)
		go func(url string) {
			log.Printf("fetching reviews on url %s\n", url)
			defer wg.Done()
			doc, err := goquery.NewDocument(url)
			if err != nil {
				return
			}

			reviews, _ := reviewsFromSelection(doc.Selection)
			b.Reviews = append(b.Reviews, reviews...)
			log.Printf("done fetching reviews on url %s\n", url)
		}(url)
	}

	wg.Wait()
}

func reviewsFromSelection(s *goquery.Selection) (r []Review, err error) {
	s.Find("ul.reviews").First().Find("div.review").Each(func(i int, s *goquery.Selection) {
		// Skip the first review: "Start your review of xxx"
		if i == 0 {
			return
		}

		review, _ := reviewFromSelection(s)
		r = append(r, review)
	})
	return
}

func authorFromSelection(s *goquery.Selection) (a Author, err error) {
	a.Name = s.Find("div.review-sidebar a.user-display-name").Text()
	a.Location = s.Find("div.review-sidebar li.user-location > b").Text()
	a.FriendCount, err = strconv.Atoi(s.Find("div.review-sidebar li.friend-count b").Text())
	a.ReviewCount, err = strconv.Atoi(s.Find("div.review-sidebar li.review-count b").Text())
	return
}

func reviewFromSelection(s *goquery.Selection) (r Review, err error) {
	r.ID = s.AttrOr("data-review-id", "")
	r.Author, err = authorFromSelection(s)
	r.Description = s.Find("div.review-content > p").Text()

	rating := s.Find(`div.review-content meta`).Eq(0).AttrOr("content", "0")
	r.Rating, err = strconv.ParseFloat(rating, 64)

	date := s.Find(`div.review-content meta`).Eq(1).AttrOr("content", "")
	r.Date, err = time.Parse("2006-01-02", date)
	return
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
