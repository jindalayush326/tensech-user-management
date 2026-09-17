package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"tensechassignment/internal/model"
	"tensechassignment/internal/repository"
)

var expectedHeader = []string{
	"email",
	"first_name",
	"last_name",
	"department",
	"status",
}

const batchSize = 200

type CSVService struct {
	repo repository.UserRepository
}

func NewCSVService(repo repository.UserRepository) *CSVService {
	return &CSVService{
		repo: repo,
	}
}

func (s *CSVService) ProcessCSV(
	ctx context.Context,
	tenantID string,
	input io.Reader,
) (*model.CSVUploadResult, error) {
	reader := csv.NewReader(input)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv header: %w", err)
	}

	if err := validateHeader(header); err != nil {
		return nil, err
	}

	result := &model.CSVUploadResult{
		Errors: make([]model.RowError, 0),
	}

	seenEmails := make(map[string]struct{})

	batch := make([]model.User, 0, batchSize)
	rowNumbers := make([]int, 0, batchSize)

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}

		inserted, failures, err := s.repo.CreateMany(ctx, batch)
		if err != nil {
			return err
		}

		result.SuccessCount += inserted

		for _, failure := range failures {
			result.FailureCount++

			message := failure.Err.Error()

			if failure.Err == repository.ErrDuplicateUser ||
				strings.Contains(message, repository.ErrDuplicateUser.Error()) {
				message = "duplicate email for tenant"
			}

			result.Errors = append(result.Errors, model.RowError{
				Row:     rowNumbers[failure.Index],
				Message: message,
			})
		}

		batch = batch[:0]
		rowNumbers = rowNumbers[:0]

		return nil
	}

	rowNumber := 1

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		rowNumber++

		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.RowError{
				Row:     rowNumber,
				Message: "malformed csv row",
			})
			continue
		}

		user, err := parseAndValidateRow(record, tenantID)
		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, model.RowError{
				Row:     rowNumber,
				Message: err.Error(),
			})
			continue
		}

		if _, exists := seenEmails[user.Email]; exists {
			result.FailureCount++
			result.Errors = append(result.Errors, model.RowError{
				Row:     rowNumber,
				Message: "duplicate email within uploaded file",
			})
			continue
		}

		seenEmails[user.Email] = struct{}{}

		batch = append(batch, *user)
		rowNumbers = append(rowNumbers, rowNumber)

		if len(batch) == batchSize {
			if err := flushBatch(); err != nil {
				return nil, err
			}
		}
	}

	if err := flushBatch(); err != nil {
		return nil, err
	}

	return result, nil
}

func validateHeader(header []string) error {
	if len(header) != len(expectedHeader) {
		return fmt.Errorf(
			"invalid csv header, expected columns: %s",
			strings.Join(expectedHeader, ","),
		)
	}

	for i, column := range header {
		column = strings.TrimSpace(strings.ToLower(column))

		if column != expectedHeader[i] {
			return fmt.Errorf(
				"invalid csv header, expected columns: %s",
				strings.Join(expectedHeader, ","),
			)
		}
	}

	return nil
}

func parseAndValidateRow(
	record []string,
	tenantID string,
) (*model.User, error) {
	if len(record) != len(expectedHeader) {
		return nil, fmt.Errorf(
			"expected %d columns, got %d",
			len(expectedHeader),
			len(record),
		)
	}

	email := strings.ToLower(strings.TrimSpace(record[0]))
	firstName := strings.TrimSpace(record[1])
	lastName := strings.TrimSpace(record[2])
	department := strings.TrimSpace(record[3])
	status := strings.ToLower(strings.TrimSpace(record[4]))

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return nil, fmt.Errorf("invalid email format")
	}

	if firstName == "" {
		return nil, fmt.Errorf("first_name is required")
	}

	if lastName == "" {
		return nil, fmt.Errorf("last_name is required")
	}

	if status == "" {
		status = "active"
	}

	if status != "active" && status != "inactive" {
		return nil, fmt.Errorf("status must be 'active' or 'inactive'")
	}

	return &model.User{
		ID:         uuid.NewString(),
		TenantID:   tenantID,
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		Department: department,
		Status:     status,
	}, nil
}