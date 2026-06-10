package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
)

type NotificationHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewNotificationHTTPClient(baseURL string, httpClient *http.Client) *NotificationHTTPClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &NotificationHTTPClient{baseURL: baseURL, client: httpClient}
}

type notificationListResponse struct {
	Notifications []dtos.NotificationResponse `json:"notifications"`
	Pagination    utils.PaginationData        `json:"pagination"`
}

type notificationCountResponse struct {
	Count int64 `json:"count"`
}

func (c *NotificationHTTPClient) List(userID string, filter repositories.NotificationFilter, page, limit int) ([]dtos.NotificationResponse, int64, error) {
	values := url.Values{
		"user_id": {userID},
		"page":    {strconv.Itoa(page)},
		"limit":   {strconv.Itoa(limit)},
	}
	if filter.Type != "" {
		values.Set("type", filter.Type)
	}
	if filter.IsRead != nil {
		values.Set("is_read", strconv.FormatBool(*filter.IsRead))
	}

	var resp notificationListResponse
	if err := c.doJSON(http.MethodGet, "/notifications?"+values.Encode(), nil, &resp); err != nil {
		return nil, 0, err
	}
	return resp.Notifications, resp.Pagination.Total, nil
}

func (c *NotificationHTTPClient) CountUnread(userID string) (int64, error) {
	values := url.Values{"user_id": {userID}}
	var resp notificationCountResponse
	if err := c.doJSON(http.MethodGet, "/notifications/unread-count?"+values.Encode(), nil, &resp); err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (c *NotificationHTTPClient) MarkAsRead(id, userID string) error {
	values := url.Values{"user_id": {userID}}
	return c.doJSON(http.MethodPut, "/notifications/"+id+"/read?"+values.Encode(), nil, nil)
}

func (c *NotificationHTTPClient) MarkAllAsRead(userID string) error {
	values := url.Values{"user_id": {userID}}
	return c.doJSON(http.MethodPut, "/notifications/read-all?"+values.Encode(), nil, nil)
}

func (c *NotificationHTTPClient) doJSON(method, path string, body any, out any) error {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
