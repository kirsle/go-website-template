package session

import (
	"errors"
	"net/http"

	"github.com/kirsle/go-website-template/webapp/models"
)

// CurrentUser returns the current logged in user via session cookie.
func CurrentUser(r *http.Request) (*models.User, error) {
	sess := Get(r)
	if sess.LoggedIn {
		// Did we already get the CurrentUser once before?
		ctx := r.Context()
		if user, ok := ctx.Value(CurrentUserKey).(*models.User); ok {
			return user, nil
		}

		// Load the associated user ID.
		return models.GetUser(sess.UserID)
	}

	return nil, errors.New("request session is not logged in")
}
