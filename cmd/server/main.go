package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"

	"github.com/Taik/yelp-reviews/yelp"
)

type yelpReviewRequest struct {
	URL     string              `json:"url"`
	Filters []yelp.ReviewFilter `json:"filters"`
}

type yelpReviewResponse struct {
	Status      string `json:"status"`
	Message     string `json:"msg,omitempty"`
	Rating      string `json:"rating"`
	ReviewCount int    `json:"review_count"`
}

func yelpReviewHandle(c echo.Context) (err error) {
	decoder := json.NewDecoder(c.Request().Body())
	request := &yelpReviewRequest{}
	resp := &yelpReviewResponse{
		Status: "OK",
	}

	err = decoder.Decode(request)
	if err != nil {
		resp.Status = "ERROR"
		resp.Message = err.Error()
		return c.JSON(http.StatusBadRequest, resp)
	}

	b, err := yelp.NewBusiness(request.URL)
	if err != nil {
		resp.Status = "ERROR"
		resp.Message = err.Error()
		return c.JSON(http.StatusBadRequest, resp)
	}

	b.FetchReviews()
	b.FilterReviews(request.Filters)

	resp.Rating = fmt.Sprintf("%.2f", b.CalculateRating())
	resp.ReviewCount = len(b.Reviews)

	return c.JSON(http.StatusOK, resp)
}

func main() {
	e := echo.New()
	e.Use(middleware.Recover(), middleware.Logger(), middleware.Gzip())

	e.POST("/", yelpReviewHandle)

	e.Run(standard.New(":1234"))
}
