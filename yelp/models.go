package yelp

import "time"

// Author defines a reviewer.
type Author struct {
	ID          string
	Name        string
	Location    string
	FriendCount int
	ReviewCount int
}

// Address defines a business address.
type Address struct {
	ID            string
	StreetAddress string
	Locality      string
	Region        string
	PostalCode    string
}

// Review defines metadata revolving a review.
type Review struct {
	ID          string
	Author      Author
	Rating      float64
	Date        time.Time
	Description string
}

// LocalBusiness defines a business.
type LocalBusiness struct {
	ID              string
	Name            string
	URL             string
	Address         Address
	AggregateRating float64
	ReviewCount     int
	Reviews         []Review
}
