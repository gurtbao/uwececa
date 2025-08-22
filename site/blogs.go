package site

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"uwece.ca/app/db"
	"uwece.ca/app/models"
	"uwece.ca/app/site/middleware"
	"uwece.ca/app/web"
)

type newBlogPageParams struct {
	Name string
}

func (s *Site) NewBlogPage(w http.ResponseWriter, r *http.Request) error {
	usr, _ := middleware.GetUser(r)
	return s.RenderTemplate(w, r, http.StatusOK, "public/new-blog", "layouts/public-base", newBlogPageParams{
		Name: usr.Name,
	})
}

type newBlogHandlerParams struct {
	Subdomain string
}

func (n *newBlogHandlerParams) From(f web.Form) error {
	name := f.Value("name")
	yearStr := f.Value("year")

	name = strings.ToLower(name)
	nameLen := len(name)
	name = strings.Map(func(r rune) rune {
		if r < 'a' || r > 'z' {
			return -1
		}

		return r
	}, name)
	if nameLen == 0 || nameLen >= 35 {
		return fmt.Errorf("Please submit a valid name (0 < len < 35).")
	}

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

	var params newBlogHandlerParams
	if err := web.FromMultipart(r, &params); err != nil {
		return s.AlertError(w, r, alertErrorParams{
			Variant: "danger",
			Message: err.Error(),
		})
	}

	_, err := models.InsertSite(r.Context(), s.db, models.NewSite{
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

	return s.RenderPlain(w, r, http.StatusOK, "public/new-blog-response", nil)
}
