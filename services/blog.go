package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"uwece.ca/app/db"
	"uwece.ca/app/models"
)

var (
	ErrSubdomainNotAvailable = errors.New("subdomain not available")
	ErrBlogDoesNotExist      = errors.New("blog does not exist")
)

type BlogService struct {
	db *db.DB
}

func NewBlogService(db *db.DB) *BlogService {
	return &BlogService{
		db: db,
	}
}

type BlogNewRequest struct {
	Name string
	Year int
}

func (b BlogNewRequest) Validate() error {
	filteredName := strings.Map(func(r rune) rune {
		if r < 'a' || r > 'z' {
			return -1
		}

		return r
	}, b.Name)
	if b.Name != filteredName {
		return errors.New("Please submit a valid name (a-z).")
	}

	if len(b.Name) == 0 || len(b.Name) >= 35 {
		return fmt.Errorf("Please submit a valid name (1 - 35 characters in length).")
	}

	if b.Year < 24 || b.Year > 30 {
		return fmt.Errorf("Please submit a valid year. (24 - 30).")
	}

	return nil
}

func (s *BlogService) New(ctx context.Context, req BlogNewRequest, usrID int) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	subdomain := fmt.Sprintf("%s.%d", req.Name, req.Year)

	_, err := models.InsertSite(ctx, s.db, models.NewSite{
		UserId:      usrID,
		Subdomain:   subdomain,
		Navbar:      "[Home](/)",
		HomeContent: "# Hello World \n Hola Mundo.",
	})
	if err != nil {
		if errors.Is(err, db.ErrUnique) {
			return ErrSubdomainNotAvailable
		}

		return fmt.Errorf("error inserting website into database: %w", err)
	}

	return nil
}

func (s *BlogService) LoadBlogFromUser(ctx context.Context, usrID int) (models.Site, error) {
	site, err := models.GetSite(ctx, s.db, db.FilterEq("user_id", usrID))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return models.Site{}, ErrBlogDoesNotExist
		}

		return models.Site{}, fmt.Errorf("error fetching site from database: %w", err)
	}

	return site, nil
}
