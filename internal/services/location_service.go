package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

type LocationService struct {
	httpClient *http.Client
	userAgent  string
}

func NewLocationService() *LocationService {
	return &LocationService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: "Obsonarium/1.0 (https://obsonarium.com)",
	}
}

// Nominatim response structures
type NominatimResult struct {
	PlaceID     int64   `json:"place_id"`
	DisplayName string  `json:"display_name"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	Address     struct {
		HouseNumber string `json:"house_number"`
		Road        string  `json:"road"`
		City        string  `json:"city"`
		State       string  `json:"state"`
		Postcode    string  `json:"postcode"`
		Country     string  `json:"country"`
	} `json:"address"`
}

type NominatimReverseResult struct {
	PlaceID     int64   `json:"place_id"`
	DisplayName string  `json:"display_name"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	Address     struct {
		HouseNumber string `json:"house_number"`
		Road        string  `json:"road"`
		City        string  `json:"city"`
		State       string  `json:"state"`
		Postcode    string  `json:"postcode"`
		Country     string  `json:"country"`
	} `json:"address"`
}

// GeocodeResult represents a geocoded address
type GeocodeResult struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	DisplayName string  `json:"display_name"`
	Address     struct {
		StreetAddress string `json:"street_address"`
		City          string `json:"city"`
		State         string `json:"state"`
		Postcode      string `json:"postcode"`
		Country       string `json:"country"`
	} `json:"address"`
}

// Geocode searches for an address using Nominatim
func (s *LocationService) Geocode(query string) ([]GeocodeResult, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("addressdetails", "1")
	params.Add("limit", "5")

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nominatim returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []NominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	geocodeResults := make([]GeocodeResult, 0, len(results))
	for _, r := range results {
		lat, err := parseFloat(r.Lat)
		if err != nil {
			continue
		}
		lon, err := parseFloat(r.Lon)
		if err != nil {
			continue
		}

		gr := GeocodeResult{
			Latitude:    lat,
			Longitude:   lon,
			DisplayName: r.DisplayName,
		}
		gr.Address.StreetAddress = fmt.Sprintf("%s %s", r.Address.HouseNumber, r.Address.Road)
		gr.Address.City = r.Address.City
		gr.Address.State = r.Address.State
		gr.Address.Postcode = r.Address.Postcode
		gr.Address.Country = r.Address.Country

		geocodeResults = append(geocodeResults, gr)
	}

	return geocodeResults, nil
}

// ReverseGeocode converts coordinates to an address using Nominatim
func (s *LocationService) ReverseGeocode(lat, lon float64) (*GeocodeResult, error) {
	baseURL := "https://nominatim.openstreetmap.org/reverse"
	params := url.Values{}
	params.Add("lat", fmt.Sprintf("%.6f", lat))
	params.Add("lon", fmt.Sprintf("%.6f", lon))
	params.Add("format", "json")

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nominatim returned status %d: %s", resp.StatusCode, string(body))
	}

	var result NominatimReverseResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	gr := &GeocodeResult{
		Latitude:    lat,
		Longitude:   lon,
		DisplayName: result.DisplayName,
	}
	gr.Address.StreetAddress = fmt.Sprintf("%s %s", result.Address.HouseNumber, result.Address.Road)
	gr.Address.City = result.Address.City
	gr.Address.State = result.Address.State
	gr.Address.Postcode = result.Address.Postcode
	gr.Address.Country = result.Address.Country

	return gr, nil
}

// OSRM response structures
type OSRMRouteResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"` // in meters
		Duration float64 `json:"duration"` // in seconds
	} `json:"routes"`
}

// Distance calculates the driving distance and ETA between two points using OSRM
func (s *LocationService) Distance(lat1, lon1, lat2, lon2 float64) (distanceKm float64, etaMinutes float64, err error) {
	// Format: lon1,lat1;lon2,lat2
	url := fmt.Sprintf("https://router.project-osrm.org/route/v1/driving/%.6f,%.6f;%.6f,%.6f?overview=false",
		lon1, lat1, lon2, lat2)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, 0, fmt.Errorf("osrm returned status %d: %s", resp.StatusCode, string(body))
	}

	var osrmResp OSRMRouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return 0, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return 0, 0, fmt.Errorf("osrm route not found")
	}

	route := osrmResp.Routes[0]
	distanceKm = route.Distance / 1000.0 // convert meters to km
	etaMinutes = route.Duration / 60.0  // convert seconds to minutes

	return distanceKm, etaMinutes, nil
}

// HaversineDistance calculates the straight-line distance between two points (in km)
func (s *LocationService) HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	deltaLat := (lat2 - lat1) * math.Pi / 180.0
	deltaLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// Helper function to parse float from string
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}


