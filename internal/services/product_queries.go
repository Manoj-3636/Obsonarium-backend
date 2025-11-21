package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type ProductQueriesService struct {
	queriesRepo  repositories.IProductQueriesRepo
	usersRepo    repositories.IUsersRepo
	emailService *EmailService
}

func NewProductQueriesService(
	queriesRepo repositories.IProductQueriesRepo,
	usersRepo repositories.IUsersRepo,
	emailService *EmailService,
) *ProductQueriesService {
	return &ProductQueriesService{
		queriesRepo:  queriesRepo,
		usersRepo:    usersRepo,
		emailService: emailService,
	}
}

func (s *ProductQueriesService) GetQueriesByRetailerID(retailerID int) ([]models.ProductQuery, error) {
	queries, err := s.queriesRepo.GetQueriesByRetailerID(retailerID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching queries by retailer ID: %w", err)
	}
	return queries, nil
}

func (s *ProductQueriesService) GetQueriesByProductID(productID int) ([]models.ProductQuery, error) {
	queries, err := s.queriesRepo.GetQueriesByProductID(productID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching queries by product ID: %w", err)
	}
	return queries, nil
}

func (s *ProductQueriesService) CreateQuery(query *models.ProductQuery) (*models.ProductQuery, error) {
	createdQuery, err := s.queriesRepo.CreateQuery(query)
	if err != nil {
		return nil, fmt.Errorf("service error creating query: %w", err)
	}
	return createdQuery, nil
}

func (s *ProductQueriesService) ResolveQuery(queryID int, responseText string) (*models.ProductQuery, error) {
	resolvedQuery, err := s.queriesRepo.ResolveQuery(queryID, responseText)
	if err != nil {
		if err == repositories.ErrProductQueryNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("service error resolving query: %w", err)
	}

	// Fetch user details to get email
	fmt.Printf("ProductQueriesService: Fetching user %d for email notification\n", resolvedQuery.User_id)
	user, err := s.usersRepo.GetUserByID(resolvedQuery.User_id)
	if err != nil {
		// Log error but don't fail the resolution
		fmt.Printf("failed to fetch user for email notification: %v\n", err)
		return resolvedQuery, nil
	}
	fmt.Printf("ProductQueriesService: Found user email: %s\n", user.Email)

	// Send email notification
	subject := fmt.Sprintf("Query resolved : %s", resolvedQuery.Query_text)
	body := fmt.Sprintf("Dear customer,\n%s", resolvedQuery.Response_text)

	if resolvedQuery.Response_text != nil {
		body = fmt.Sprintf("Dear customer,\n%s", *resolvedQuery.Response_text)
	}

	fmt.Println("ProductQueriesService: Attempting to send email...")
	err = s.emailService.SendEmail(user.Email, subject, body)
	if err != nil {
		// Log error but don't fail the resolution
		fmt.Printf("failed to send email notification: %v\n", err)
	} else {
		fmt.Println("ProductQueriesService: Email sent successfully")
	}

	return resolvedQuery, nil
}
