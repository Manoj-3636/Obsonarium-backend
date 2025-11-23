package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type ShopWithDistance struct {
	Retailer models.Retailer `json:"retailer"`
	Distance float64         `json:"distance"` // in km
	ETA      float64         `json:"eta"`     // in minutes
}

type ShopsService struct {
	retailersRepo   repositories.IRetailersRepo
	locationService *LocationService
}

func NewShopsService(retailersRepo repositories.IRetailersRepo) *ShopsService {
	return &ShopsService{
		retailersRepo:   retailersRepo,
		locationService: NewLocationService(),
	}
}

// GetNearbyShops gets retailers within a radius and calculates distance/ETA
func (s *ShopsService) GetNearbyShops(lat, lon, radiusKm float64) ([]ShopWithDistance, error) {
	// Get retailers within radius using Haversine
	retailers, err := s.retailersRepo.GetNearbyRetailers(lat, lon, radiusKm)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nearby retailers: %w", err)
	}

	shops := make([]ShopWithDistance, 0, len(retailers))
	for _, retailer := range retailers {
		if retailer.Latitude == nil || retailer.Longitude == nil {
			continue // Skip retailers without coordinates
		}

		// Calculate distance using Haversine (fast, straight-line)
		distanceKm := s.locationService.HaversineDistance(lat, lon, *retailer.Latitude, *retailer.Longitude)

		// Calculate ETA using OSRM (driving route)
		etaMinutes := 0.0
		_, eta, err := s.locationService.Distance(lat, lon, *retailer.Latitude, *retailer.Longitude)
		if err == nil {
			etaMinutes = eta
		}

		shops = append(shops, ShopWithDistance{
			Retailer: retailer,
			Distance: distanceKm,
			ETA:      etaMinutes,
		})
	}

	return shops, nil
}


