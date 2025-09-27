package bidder

import (
	"auction-simulator/pkg/models"
	"math/rand"
	"time"
)

// TODO: Strategy algorithm for the user for bidding
func Bidder(bidderId int, details models.AuctionDetails) *models.Bid {
	// Bidding delay
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

	// 80% chance to bid
	if rand.Intn(100) < 80 {
		bidAmount := float64(details.AuctionId)*100 + rand.Float64()*50
		return &models.Bid{
			BidderId:  bidderId,
			Amount:    bidAmount,
			Timestamp: time.Now(),
		}
	}

	// No bid
	return nil
}
