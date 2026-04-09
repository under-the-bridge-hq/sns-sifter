package xapi

import (
	"encoding/json"
	"fmt"
)

type FollowingPage struct {
	Users     []User
	NextToken string
}

func (c *Client) GetFollowingPage(userID string, maxResults int, paginationToken string) (*FollowingPage, error) {
	params := map[string]string{
		"user.fields": "id,username,name,description,public_metrics",
		"max_results": fmt.Sprintf("%d", maxResults),
	}
	if paginationToken != "" {
		params["pagination_token"] = paginationToken
	}

	body, err := c.get("/2/users/"+userID+"/following", params)
	if err != nil {
		return nil, err
	}

	var resp UsersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &FollowingPage{
		Users:     resp.Data,
		NextToken: resp.Meta.NextToken,
	}, nil
}

type FollowingCallback func(page []User, apiCalls int)

func (c *Client) GetAllFollowing(userID string, cb FollowingCallback) ([]User, int, error) {
	var all []User
	apiCalls := 0
	token := ""

	for {
		page, err := c.GetFollowingPage(userID, 1000, token)
		if err != nil {
			return all, apiCalls, err
		}
		apiCalls++
		all = append(all, page.Users...)

		if cb != nil {
			cb(page.Users, apiCalls)
		}

		if page.NextToken == "" {
			break
		}
		token = page.NextToken
	}

	return all, apiCalls, nil
}
