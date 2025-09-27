package auction

import (
	"auction-simulator/pkg/bidder"
	"auction-simulator/pkg/models"
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	NumAuctions    = 40
	SimulatedVCPUs = 4
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

func StartAuctions() {
	runtime.GOMAXPROCS(SimulatedVCPUs)

	overallStartTime := time.Now()

	var wg sync.WaitGroup
	resultChan := make(chan models.AuctionResult, NumAuctions)

	fmt.Printf("Launching %d concurrent auctions...\n", NumAuctions)
	for i := 1; i <= NumAuctions; i++ {
		wg.Add(1)
		go startAuction(i, &wg, resultChan)
	}

	wg.Wait()

	close(resultChan)

	overallEndTime := time.Now()
	totalExecutionTime := overallEndTime.Sub(overallStartTime)

	var completedAuctions int
	for res := range resultChan {
		completedAuctions++
		status := "No Bids Received"
		if res.WinnerId != 0 {
			status = fmt.Sprintf("Winner: Bidder %d %v", res.WinnerId, res.WinningBid)
		}

		fmt.Printf("Auction %d finished in %s. %s\n", res.AuctionId, res.Duration.String(), status)
	}

	fmt.Printf("Total Auctions Completed: %d\n", completedAuctions)
	fmt.Printf("Overall Execution Time: %s\n", totalExecutionTime)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\n--- Standardized Resource Metrics (Go Runtime) ---\n")
	fmt.Printf("vCPU Standard (GOMAXPROCS): %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("RAM Standard (Heap Alloc): %.2f MB\n", float64(m.HeapAlloc)/1024/1024)
	fmt.Printf("Goroutines running: %d\n", runtime.NumGoroutine())
}

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
