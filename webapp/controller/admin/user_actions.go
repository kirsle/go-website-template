package admin

import (
	"net/http"
	"strconv"

	"github.com/aichaos/silhouette/webapp/models"
	"github.com/aichaos/silhouette/webapp/models/deletion"
	"github.com/aichaos/silhouette/webapp/session"
	"github.com/aichaos/silhouette/webapp/templates"
)

// Admin actions against a user account.
func UserActions() http.HandlerFunc {
	tmpl := templates.Must("admin/user_actions.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			intent  = r.FormValue("intent")
			confirm = r.Method == http.MethodPost
			userId  uint64
		)

		if idInt, err := strconv.Atoi(r.FormValue("user_id")); err == nil {
			userId = uint64(idInt)
		} else {
			session.FlashError(w, r, "Invalid or missing user_id parameter: %s", err)
			templates.Redirect(w, "/admin")
			return
		}

		// Get this user.
		user, err := models.GetUser(userId)
		if err != nil {
			session.FlashError(w, r, "Didn't find user ID in database: %s", err)
			templates.Redirect(w, "/admin")
			return
		}

		switch intent {
		case "ban":
			if confirm {
				status := r.PostFormValue("status")
				if status == "active" {
					user.Status = models.UserStatusActive
				} else if status == "banned" {
					user.Status = models.UserStatusBanned
				}

				user.Save()
				session.Flash(w, r, "User ban status updated!")
				templates.Redirect(w, "/u/"+user.Username)
				return
			}
		case "promote":
			if confirm {
				action := r.PostFormValue("action")
				user.IsAdmin = action == "promote"
				user.Save()
				session.Flash(w, r, "User admin status updated!")
				templates.Redirect(w, "/u/"+user.Username)
				return
			}
		case "delete":
			if confirm {
				if err := deletion.DeleteUser(user); err != nil {
					session.FlashError(w, r, "Failed when deleting the user: %s", err)
				} else {
					session.Flash(w, r, "User has been deleted!")
				}
				templates.Redirect(w, "/admin")
				return
			}
		default:
			session.FlashError(w, r, "Unsupported admin user intent: %s", intent)
			templates.Redirect(w, "/admin")
			return
		}

		var vars = map[string]interface{}{
			"Intent": intent,
			"User":   user,
		}
		if err := tmpl.Execute(w, r, vars); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
