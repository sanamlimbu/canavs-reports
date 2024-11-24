package canvas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type User struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	SortableName  string    `json:"sortable_name"`
	ShortName     string    `json:"short_name"`
	SISUserID     string    `json:"sis_user_id"`
	IntegrationID string    `json:"intergration_id"`
	SISImportID   int       `json:"sis_import_id"`
	LoginID       string    `json:"login_id"`
}

// GetUserByID retrieves user with given ID.
func (c *CanvasClient) GetUserByID(userID int) (User, int, error) {
	requestUrl := fmt.Sprintf("%s/users/%d", c.baseUrl, userID)

	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return User{}, http.StatusInternalServerError, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return User{}, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return User{}, res.StatusCode, fmt.Errorf("error fetching user: %d", userID)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return User{}, http.StatusInternalServerError, err
	}

	var user User

	if err := json.Unmarshal(body, &user); err != nil {
		return user, http.StatusInternalServerError, err
	}

	return user, http.StatusOK, nil
}
