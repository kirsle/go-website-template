package account

import (
	"net/http"

	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/models"
	"github.com/aichaos/silhouette/webapp/session"
	"github.com/aichaos/silhouette/webapp/templates"
)

// Search controller.
func Search() http.HandlerFunc {
	tmpl := templates.Must("account/search.html")

	// Whitelist for ordering options.
	var sortWhitelist = []string{
		"last_login_at desc",
		"created_at desc",
		"username",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Search filters.
		var (
			username = r.FormValue("username") // email or username
			sort     = r.FormValue("sort")
			sortOK   bool
		)

		// Get current user.
		currentUser, err := session.CurrentUser(r)
		if err != nil {
			session.FlashError(w, r, "Couldn't get current user!")
			templates.Redirect(w, "/")
			return
		}

		// Sort options.
		for _, v := range sortWhitelist {
			if sort == v {
				sortOK = true
				break
			}
		}
		if !sortOK {
			sort = "last_login_at desc"
		}

		pager := &models.Pagination{
			PerPage: config.PageSizeMemberSearch,
			Sort:    sort,
		}
		pager.ParsePage(r)

		users, err := models.SearchUsers(currentUser, &models.UserSearch{
			EmailOrUsername: username,
		}, pager)
		if err != nil {
			session.FlashError(w, r, "Couldn't search users: %s", err)
		}

		var vars = map[string]interface{}{
			"Users": users,
			"Pager": pager,

			// Search filter values.
			"EmailOrUsername": username,
			"Sort":            sort,
		}

		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
