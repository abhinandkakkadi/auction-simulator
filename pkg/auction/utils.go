package auction

import (
	"auction-simulator/pkg/models"
	"fmt"
)

const NumAttributes = 20

func GenerateAuctionDetails(id int) models.AuctionDetails {
	details := models.AuctionDetails{AuctionId: id}
	for i := range NumAttributes {
		details.Attributes[i] = models.Attribute{
			Name:        fmt.Sprintf("Attribute-%d", i+1),
			Description: fmt.Sprintf("Description-%d", i+1),
		}
	}

	return details
}
