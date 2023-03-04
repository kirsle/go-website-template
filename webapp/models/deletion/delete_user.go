package deletion

import (
	"fmt"

	"github.com/aichaos/silhouette/webapp/log"
	"github.com/aichaos/silhouette/webapp/models"
)

// DeleteUser wipes a user and all associated data from the database.
func DeleteUser(user *models.User) error {
	log.Error("BEGIN DeleteUser(%d, %s)", user.ID, user.Username)

	// Remove all linked tables and assets.
	type remover struct {
		Step string
		Fn   func(uint64) error
	}

	var todo = []remover{
		// e.g.
		// {"Notifications", func(userID uint64) error},
		// {"Likes", DeleteLikes},
		// {"Threads", DeleteForumThreads},
	}
	for _, item := range todo {
		if err := item.Fn(user.ID); err != nil {
			return fmt.Errorf("%s: %s", item.Step, err)
		}
	}

	// Remove the user itself.
	return user.Delete()
}
