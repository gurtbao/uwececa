package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"uwece.ca/app/config"
	"uwece.ca/app/db"
	"uwece.ca/app/mailer"
	"uwece.ca/app/models"
	"uwece.ca/app/utils"
	"uwece.ca/app/web"
)

var (
	ErrValidationFailed    = errors.New("Validation Failed")
	ErrUserExists          = errors.New("user already exists")
	ErrUserDoesNotExist    = errors.New("user does not exist")
	ErrUserNotVerified     = errors.New("user not verified")
	ErrUserWrongPassword   = errors.New("wrong password for user")
	ErrTokenNotFound       = errors.New("verification token not found")
	ErrTokenExpired        = errors.New("verification token expired")
	ErrSessionExpired      = errors.New("user session expired")
	ErrSessionDoesNotExist = errors.New("user session does not exist")
)

type UserService struct {
	db     *db.DB
	mailer mailer.Mailer
	config *config.Config
}

func NewUserService(db *db.DB, mailer mailer.Mailer, config *config.Config) *UserService {
	return &UserService{
		db:     db,
		mailer: mailer,
		config: config,
	}
}

func (s UserService) GetEmail(netID string) string {
	return fmt.Sprintf("%s@%s", netID, s.config.Core.EmailDomain)
}

type UserSignupRequest struct {
	NetID           string
	Password        string
	PasswordConfirm string
	Name            string
}

func (s UserSignupRequest) Validate() error {
	if len(s.NetID) == 0 || len(s.NetID) > 35 {
		return errors.New("Please provide a valid netID (1 - 35 characters in length).")
	}

	filteredNetID := strings.Map(func(r rune) rune {
		if r < '0' || 'z' < r {
			return -1
		}

		return r
	}, s.NetID)

	if filteredNetID != s.NetID {
		return errors.New("Please provide a valid netID (1-9,a-z).")
	}

	if s.Name == "" {
		return errors.New("Please provide a name :).")
	}

	if len(s.Password) < 12 {
		return errors.New("Please provide a password of length 12 or greater.")
	}

	if s.Password != s.PasswordConfirm {
		return errors.New("Password and password confirmation must match.")
	}

	return nil
}

type UserSignupResponse struct {
	Email string
	Name  string
}

func (s *UserService) Signup(ctx context.Context, req UserSignupRequest) (UserSignupResponse, error) {
	if err := req.Validate(); err != nil {
		return UserSignupResponse{}, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	hashedPassword := utils.HashPassword(req.Password)

	usr, err := models.InsertUser(ctx, s.db, models.NewUser{
		NetID:    req.NetID,
		Password: hashedPassword,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, db.ErrUnique) {
			return UserSignupResponse{}, ErrUserExists
		}

		return UserSignupResponse{}, fmt.Errorf("failed to create user: %s", err)
	}

	e, err := models.InsertEmail(ctx, s.db, models.NewEmail{
		Token:   utils.NewToken(),
		UserId:  usr.Id,
		Expires: time.Now().Add(48 * time.Hour),
	})
	if err != nil {
		return UserSignupResponse{}, fmt.Errorf("error creating verification email in database: %w", err)
	}

	if err := s.mailer.SendVerificationEmail(s.GetEmail(usr.NetID), usr.Name, e.Token); err != nil {
		return UserSignupResponse{}, fmt.Errorf("error sending verification email: %w", err)
	}

	return UserSignupResponse{
		Email: s.GetEmail(usr.NetID),
		Name:  usr.Name,
	}, nil
}

type UserLoginRequest struct {
	NetID    string
	Password string
}

func (r UserLoginRequest) Validate() error {
	if r.NetID == "" {
		return errors.New("Please provide a non-zero NetID.")
	}

	if r.Password == "" {
		return errors.New("Please provide a valid password.")
	}

	return nil
}

type UserLoginResponse struct {
	Session web.Session
}

func (s *UserService) Login(ctx context.Context, req UserLoginRequest) (UserLoginResponse, error) {
	if err := req.Validate(); err != nil {
		return UserLoginResponse{}, fmt.Errorf("%w, %v", ErrValidationFailed, err)
	}

	usr, err := models.GetUser(ctx, s.db, db.FilterEq("net_id", req.NetID))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return UserLoginResponse{}, ErrUserDoesNotExist
		}

		return UserLoginResponse{}, fmt.Errorf("error fetching user from database: %w", err)
	}

	if usr.VerifiedAt == nil {
		return UserLoginResponse{}, ErrUserNotVerified
	}

	ok, err := utils.VerifyPassword(req.Password, usr.Password)
	if err != nil {
		return UserLoginResponse{}, fmt.Errorf("error verifiying password: %w", err)
	}

	if !ok {
		return UserLoginResponse{}, ErrUserWrongPassword
	}

	session := web.NewSession()
	_, err = models.InsertSession(ctx, s.db, models.NewSession{
		Token:   session.Token,
		Expires: session.Expiry,
		UserId:  usr.Id,
	})
	if err != nil {
		return UserLoginResponse{}, fmt.Errorf("error inserting session token: %w", err)
	}

	return UserLoginResponse{Session: session}, nil
}

func (s *UserService) Verify(ctx context.Context, token utils.Token) error {
	e, err := models.GetEmail(ctx, s.db, db.FilterEq("token", token))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("error fetching email token from database: %w", err)
	}

	if time.Now().After(e.Expires) {
		return ErrTokenExpired
	}

	updates := db.Updates(
		db.Update("updated_at", time.Now()),
		db.Update("verified_at", time.Now()),
	)

	if err := models.UpdateUser(ctx, s.db, updates, db.FilterEq("id", e.UserId)); err != nil {
		return fmt.Errorf("error setting user as verified: %w", err)
	}

	return nil
}

func (s *UserService) LoadSession(ctx context.Context, token utils.Token) (models.User, error) {
	dbs, err := models.GetSession(ctx, s.db, db.FilterEq("token", token))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return models.User{}, ErrSessionDoesNotExist
		}

		return models.User{}, fmt.Errorf("error loading session from database: %w", err)
	}

	if time.Now().After(dbs.Expires) {
		return models.User{}, ErrSessionExpired
	}

	usr, err := models.GetUser(ctx, s.db, db.FilterEq("id", dbs.UserId))
	if err != nil {
		if errors.Is(err, db.ErrNoRows) {
			return models.User{}, ErrUserDoesNotExist
		}
		return models.User{}, fmt.Errorf("failed to fetch user while loading session: %w", err)
	}

	return usr, nil
}
