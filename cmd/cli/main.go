package main

import (
	"fmt"
	"strings"

	"github.com/Taik/yelp-reviews/yelp"
)

func main() {
	url := "http://www.yelp.com/biz/sal-kris-and-charlies-deli-astoria"
	business, _ := yelp.NewBusiness(url)

	business.FetchReviews()
	var reviews []yelp.Review

	uniqueLocations := map[string]int{}
	var sum float64

	for _, r := range business.Reviews {
		reviews = append(reviews, r)
		sum += r.Rating

		loc := strings.TrimSpace(strings.ToLower(r.Author.Location))
		uniqueLocations[loc]++
	}

	fmt.Printf("Total reviews: %d, rating: %f, unique locations: %d\n",
		len(reviews),
		sum/float64(len(reviews)),
		len(uniqueLocations),
	)
}
