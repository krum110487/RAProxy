package server

type ReviewersInfo struct {
	Reviewers []Reviewer `json:"reviewers"`
}

type Reviewer struct {
	UserID        string   `json:"userId"`
	Username      string   `json:"username"`
	Location      string   `json:"location"`
	FavoriteGames []string `json:"favoriteGames"`
}

func (ri *ReviewersInfo) GetReviewerByUserID(userID string) *Reviewer {
	for i := range ri.Reviewers {
		if ri.Reviewers[i].UserID == userID {
			return &ri.Reviewers[i]
		}
	}
	return nil
}
