package models

import "time"

type Attribute struct {
	Name        string
	Description string
}

type Bid struct {
	BidderId  int
	Amount    float64
	Timestamp time.Time
}

type AuctionDetails struct {
	AuctionId  int
	Attributes [20]Attribute
}

type AuctionResult struct {
	AuctionId  int
	WinnerId   int
	WinningBid float64
	Duration   time.Duration
}
