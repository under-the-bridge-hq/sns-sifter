package xapi

type User struct {
	ID                string        `json:"id"`
	Username          string        `json:"username"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	PublicMetrics     PublicMetrics `json:"public_metrics"`
}

type PublicMetrics struct {
	FollowersCount int `json:"followers_count"`
	FollowingCount int `json:"following_count"`
	TweetCount     int `json:"tweet_count"`
	ListedCount    int `json:"listed_count"`
	LikeCount      int `json:"like_count"`
}

type UsersResponse struct {
	Data []User         `json:"data"`
	Meta ResponseMeta   `json:"meta"`
}

type UserResponse struct {
	Data User `json:"data"`
}

type ResponseMeta struct {
	ResultCount   int    `json:"result_count"`
	NextToken     string `json:"next_token"`
}

type APIError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status int    `json:"status"`
}

func (e *APIError) Error() string {
	return e.Title + ": " + e.Detail
}
