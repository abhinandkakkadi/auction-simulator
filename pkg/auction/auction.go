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

func startAuction(auctionId int, wg *sync.WaitGroup, resultChan chan models.AuctionResult) {
	defer wg.Done()

	var bidWg sync.WaitGroup
	bidChan, closeSignal := StartBidding(auctionId, &bidWg)

	// Set auction timeout
	timer := time.NewTimer(AuctionTimeout)
	startTime := time.Now()
	var receivedBids []*models.Bid

	bidCollectorDone := make(chan struct{})
	go func() {
		defer close(bidCollectorDone)
		for {
			select {
			case bid := <-bidChan:
				receivedBids = append(receivedBids, bid)
			case <-timer.C:
				return // Timeout reached
			}
		}
	}()

	<-bidCollectorDone

	close(closeSignal) // Stop accepting bids
	bidWg.Wait()       // Make sure all bidders exit

	duration := time.Since(startTime)
	var winnerId int
	var winningBid float64 = -1.0

	for _, bid := range receivedBids {
		if bid.Amount > winningBid {
			winningBid = bid.Amount
			winnerId = bid.BidderId
		}
	}

	resultChan <- models.AuctionResult{
		AuctionId:  auctionId,
		WinnerId:   winnerId,
		WinningBid: winningBid,
		Duration:   duration,
	}
}

func StartAuctions() {}

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
