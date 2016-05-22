package yelp

import (
	"fmt"
	"strings"
	"time"
)

// Author defines a reviewer.
type Author struct {
	ID          string `csss:"a.user-display-name;attr=data-hovercard-id"`
	Name        string `csss:"a.user-display-name;text"`
	Location    string `csss:"a.user-location;text"`
	FriendCount int    `csss:"li.friend-count > b;text"`
	ReviewCount int    `csss:"li.review-count > b;text"`
}

// Address defines a business address.
type Address struct {
	StreetAddress string `csss:"span:nth-child(1);text"`
	Locality      string `csss:"span:nth-child(3);text"`
	Region        string `csss:"span:nth-child(4);text"`
	PostalCode    string `csss:"span:nth-child(5);text"`
}

// Review defines metadata revolving a review.
type Review struct {
	ID          string  `csss:";attr=data-review-id"`
	Author      Author  `csss:"div.ypassport;obj"`
	Rating      float64 `csss:"meta[itemprop=ratingValue];attr=content"`
	DateStr     string  `csss:"meta[itemprop=datePublished];attr=content"`
	Date        time.Time
	Description string `csss:"p[itemprop=description];text"`
}

// XXX: SqrapePostFlight defines custom parsing logic for scraped fields.
func (r *Review) SqrapePostFlight(context ...interface{}) (err error) {
	if r.DateStr != "" {
		r.Date, err = time.Parse("2006-01-02", r.DateStr)
	}
	return err
}

// LocalBusiness defines a business.
type LocalBusiness struct {
	ID              string `csss:"meta[name='yelp-biz-id'];attr=content"`
	Name            string `csss:"h1.biz-page-title;text"`
	URL             string
	Address         Address  `csss:"address;obj"`
	AggregateRating float64  `csss:"div.biz-rating meta[itemprop=ratingValue];attr=content"`
	ReviewCount     int      `csss:"div.biz-rating span[itemprop=reviewCount];text"`
	Reviews         []Review `csss:"div.review;obj"`
}

// XXX: SqrapeFieldSelect skips parsing other fields when parsing paginated fields.
func (b *LocalBusiness) SqrapeFieldSelect(fieldName string, context ...interface{}) (bool, error) {
	if len(context) != 1 {
		return false, fmt.Errorf("invalid context count")
	}
	// 1st field: is pagination
	if context[0].(bool) == true {
		return fieldName == "Reviews", nil
	}
	return true, nil
}

// XXX: SqrapePostFlight defines custom parsing logic for scraped fields.
func (b *LocalBusiness) SqrapePostFlight(context ...interface{}) (err error) {
	b.Name = strings.TrimSpace(b.Name)
	return err
}

// ReviewFilter defines filter data to filter reviews by.
type ReviewFilter struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
