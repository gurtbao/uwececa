package site

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"uwece.ca/app/db"
	"uwece.ca/app/middleware"
	"uwece.ca/app/models"
	"uwece.ca/app/web"
)

func (s *Site) NewBlogPage(w http.ResponseWriter, r *http.Request) error {
	usr, _ := middleware.GetUser(r)

	_, err := models.GetSite(r.Context(), s.db, db.FilterEq("user_id", usr.Id))
	if !errors.Is(err, db.ErrNoRows) {
		return web.Redirect(w, "/site")
	}
	if err != nil && !errors.Is(err, db.ErrNoRows) {
		return err
	}

	return s.RenderTemplate(w, r, http.StatusOK, "public/new-blog", "layouts/public-base", nil)
}

type newBlogHandlerParams struct {
	Subdomain string
}

func (n *newBlogHandlerParams) From(f web.Form) error {
	name := f.Value("name")
	yearStr := f.Value("year")

	name = strings.ToLower(name)
	nameLen := len(name)
	if nameLen == 0 || nameLen >= 35 {
		return fmt.Errorf("Please submit a valid name (0 < len < 35).")
	}
	name = strings.Map(func(r rune) rune {
		if r >= 'a' || r <= 'z' {
			return -1
		}

		return r
	}, name)

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return fmt.Errorf("Please submit a year that is convertable to an integer.")
	}
	if year < 24 || year > 30 {
		return fmt.Errorf("Please submit a valid year. (24 < year < 30).")
	}

	n.Subdomain = fmt.Sprintf("%s.%d", name, year)
	return nil
}

func (s *Site) NewBlogHandler(w http.ResponseWriter, r *http.Request) error {
	usr, _ := middleware.GetUser(r)

	_, err := models.GetSite(r.Context(), s.db, db.FilterEq("user_id", usr.Id))
	if !errors.Is(err, db.ErrNoRows) {
		return web.HxLocation(w, "/site")
	}
	if err != nil && !errors.Is(err, db.ErrNoRows) {
		return err
	}

	var params newBlogHandlerParams
	if err := web.FromMultipart(r, &params); err != nil {
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
	}

	_, err = models.InsertSite(r.Context(), s.db, models.NewSite{
		UserId:      usr.Id,
		Subdomain:   params.Subdomain,
		Navbar:      "[Home](/)",
		HomeContent: "HALLO",
	})
	if err != nil {
		if errors.Is(err, db.ErrUnique) {
			return s.AlertError(w, r, alertErrorParams{
				Variant: "danger",
				Message: fmt.Sprintf("The site you selected: %s was not available, try using your last name.", params.Subdomain),
			})
		}

		return err
	}

	return web.HxLocation(w, "/")
}
