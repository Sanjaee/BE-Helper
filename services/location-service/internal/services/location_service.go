package services

import (
	"fmt"
	"location-service/internal/models"
	"location-service/internal/publisher"
	"location-service/internal/repository"
	"math"
	"time"

	"github.com/google/uuid"
)

type LocationService interface {
	UpdateLocation(req *models.UpdateLocationRequest) (*models.LocationResponse, error)
	GetOrderLocation(orderID uuid.UUID) (*models.LocationResponse, error)
	GetLocationHistory(orderID uuid.UUID) ([]models.LocationHistoryResponse, error)
}

type locationService struct {
	locationRepo   repository.LocationRepository
	eventPublisher *publisher.EventPublisher
}

func NewLocationService(locationRepo repository.LocationRepository, eventPublisher *publisher.EventPublisher) LocationService {
	return &locationService{
		locationRepo:   locationRepo,
		eventPublisher: eventPublisher,
	}
}

func (s *locationService) UpdateLocation(req *models.UpdateLocationRequest) (*models.LocationResponse, error) {
	// Check if tracking already exists
	existingTracking, err := s.locationRepo.GetTrackingByOrderID(req.OrderID)
	if err != nil {
		// Create new tracking if it doesn't exist
		tracking := &models.OrderTracking{
			OrderID:           req.OrderID,
			ServiceProviderID: req.ServiceProviderID,
			CurrentLatitude:   req.Latitude,
			CurrentLongitude:  req.Longitude,
			TrackingStatus:    "ACTIVE",
			LastUpdated:       time.Now(),
		}

		if err := s.locationRepo.CreateTracking(tracking); err != nil {
			return nil, fmt.Errorf("failed to create tracking: %w", err)
		}

		// Create location history entry
		history := &models.LocationHistory{
			OrderID:           req.OrderID,
			ServiceProviderID: req.ServiceProviderID,
			Latitude:          req.Latitude,
			Longitude:         req.Longitude,
			SpeedKmh:          req.SpeedKmh,
			AccuracyMeters:    req.AccuracyMeters,
			HeadingDegrees:    req.HeadingDegrees,
			RecordedAt:        time.Now(),
		}

		if err := s.locationRepo.CreateLocationHistory(history); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to create location history: %v\n", err)
		}

		// Publish location updated event
		if err := s.eventPublisher.PublishLocationUpdated(tracking); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish location updated event: %v\n", err)
		}

		return &models.LocationResponse{
			OrderID:                 tracking.OrderID,
			ServiceProviderID:       tracking.ServiceProviderID,
			Latitude:                tracking.CurrentLatitude,
			Longitude:               tracking.CurrentLongitude,
			DistanceKm:              tracking.DistanceKm,
			EstimatedArrivalMinutes: tracking.EstimatedArrivalMinutes,
			TrackingStatus:          tracking.TrackingStatus,
			LastUpdated:             tracking.LastUpdated,
		}, nil
	}

	// Update existing tracking
	existingTracking.CurrentLatitude = req.Latitude
	existingTracking.CurrentLongitude = req.Longitude
	existingTracking.LastUpdated = time.Now()

	// Calculate distance (simplified calculation)
	// In a real implementation, you would calculate distance to destination
	existingTracking.DistanceKm = s.calculateDistance(req.Latitude, req.Longitude, 0, 0) // Placeholder

	// Calculate estimated arrival time (simplified)
	// In a real implementation, you would calculate based on current speed and distance
	existingTracking.EstimatedArrivalMinutes = int(existingTracking.DistanceKm * 2) // Placeholder

	if err := s.locationRepo.UpdateTracking(existingTracking); err != nil {
		return nil, fmt.Errorf("failed to update tracking: %w", err)
	}

	// Create location history entry
	history := &models.LocationHistory{
		OrderID:           req.OrderID,
		ServiceProviderID: req.ServiceProviderID,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		SpeedKmh:          req.SpeedKmh,
		AccuracyMeters:    req.AccuracyMeters,
		HeadingDegrees:    req.HeadingDegrees,
		RecordedAt:        time.Now(),
	}

	if err := s.locationRepo.CreateLocationHistory(history); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create location history: %v\n", err)
	}

	// Publish location updated event
	if err := s.eventPublisher.PublishLocationUpdated(existingTracking); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish location updated event: %v\n", err)
	}

	return &models.LocationResponse{
		OrderID:                 existingTracking.OrderID,
		ServiceProviderID:       existingTracking.ServiceProviderID,
		Latitude:                existingTracking.CurrentLatitude,
		Longitude:               existingTracking.CurrentLongitude,
		DistanceKm:              existingTracking.DistanceKm,
		EstimatedArrivalMinutes: existingTracking.EstimatedArrivalMinutes,
		TrackingStatus:          existingTracking.TrackingStatus,
		LastUpdated:             existingTracking.LastUpdated,
	}, nil
}

func (s *locationService) GetOrderLocation(orderID uuid.UUID) (*models.LocationResponse, error) {
	tracking, err := s.locationRepo.GetTrackingByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("tracking not found: %w", err)
	}

	return &models.LocationResponse{
		OrderID:                 tracking.OrderID,
		ServiceProviderID:       tracking.ServiceProviderID,
		Latitude:                tracking.CurrentLatitude,
		Longitude:               tracking.CurrentLongitude,
		DistanceKm:              tracking.DistanceKm,
		EstimatedArrivalMinutes: tracking.EstimatedArrivalMinutes,
		TrackingStatus:          tracking.TrackingStatus,
		LastUpdated:             tracking.LastUpdated,
	}, nil
}

func (s *locationService) GetLocationHistory(orderID uuid.UUID) ([]models.LocationHistoryResponse, error) {
	history, err := s.locationRepo.GetLocationHistory(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location history: %w", err)
	}

	var response []models.LocationHistoryResponse
	for _, h := range history {
		response = append(response, models.LocationHistoryResponse{
			ID:                h.ID,
			OrderID:           h.OrderID,
			ServiceProviderID: h.ServiceProviderID,
			Latitude:          h.Latitude,
			Longitude:         h.Longitude,
			SpeedKmh:          h.SpeedKmh,
			AccuracyMeters:    h.AccuracyMeters,
			HeadingDegrees:    h.HeadingDegrees,
			RecordedAt:        h.RecordedAt,
		})
	}

	return response, nil
}

// calculateDistance calculates the distance between two coordinates using Haversine formula
func (s *locationService) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
