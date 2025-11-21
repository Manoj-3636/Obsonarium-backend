package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type RetailerAddressesService struct {
	addressesRepo repositories.IRetailerAddressesRepo
	retailersRepo repositories.IRetailersRepo
}

func NewRetailerAddressesService(addressesRepo repositories.IRetailerAddressesRepo, retailersRepo repositories.IRetailersRepo) *RetailerAddressesService {
	return &RetailerAddressesService{
		addressesRepo: addressesRepo,
		retailersRepo: retailersRepo,
	}
}

func (s *RetailerAddressesService) CreateAddress(email string, address *models.RetailerAddress) error {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get retailer: %w", err)
	}

	address.Retailer_id = retailer.Id
	return s.addressesRepo.CreateAddress(address)
}

func (s *RetailerAddressesService) GetAddresses(email string) ([]models.RetailerAddress, error) {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get retailer: %w", err)
	}

	return s.addressesRepo.GetAddressesByRetailerID(retailer.Id)
}

func (s *RetailerAddressesService) DeleteAddress(email string, addressID int) error {
	// Verify the address belongs to this retailer
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get retailer: %w", err)
	}

	address, err := s.addressesRepo.GetAddress(addressID)
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	if address.Retailer_id != retailer.Id {
		return fmt.Errorf("address does not belong to retailer")
	}

	return s.addressesRepo.DeleteAddress(addressID)
}
