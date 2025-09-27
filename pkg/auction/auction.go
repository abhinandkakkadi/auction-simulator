package auction

import (
	"auction-simulator/pkg/bidder"
	"auction-simulator/pkg/models"
	"sync"
	"time"
)

const (
	NumBidders     = 100
	AuctionTimeout = 500 * time.Millisecond
)

func StartBidding(auctionId int, bidWg *sync.WaitGroup) (chan *models.Bid, chan struct{}) {
	auctionDetails := GenerateAuctionDetails(auctionId)

	bidChan := make(chan *models.Bid, NumBidders)
	closeSignal := make(chan struct{}) // Auction closed signal

	for i := 1; i <= NumBidders; i++ {
		bidWg.Add(1)
		go func(id int) {
			defer bidWg.Done()
			select {
			case <-closeSignal:
				return
			default:
				if bid := bidder.Bidder(id, auctionDetails); bid != nil {
					select {
					case bidChan <- bid:
					case <-closeSignal:
						// Bid was generated just as the auction closed
					}
				}
			}
		}(i)
	}

	return bidChan, closeSignal
}
