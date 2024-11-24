package canvas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/guregu/null/v5"
)

type Account struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	ParentAccountID null.Int `json:"parent_account_id"`
	RootAccountID   null.Int `json:"root_account_id"`
	WorkflowState   string   `json:"workflow_state"`
}

// GetAccountByID retrieves account of given ID.
func (c *CanvasClient) GetAccountByID(accountID int) (Account, int, error) {
	requestUrl := fmt.Sprintf("%s/accounts/%d", c.baseUrl, accountID)

	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return Account{}, http.StatusInternalServerError, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Account{}, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Account{}, res.StatusCode, fmt.Errorf("error fetching account: %d", accountID)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Account{}, http.StatusInternalServerError, err
	}

	var account Account

	if err := json.Unmarshal(body, &account); err != nil {
		return Account{}, http.StatusInternalServerError, err
	}

	return account, http.StatusOK, nil
}
