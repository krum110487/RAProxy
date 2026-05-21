package server

type GameInfo struct {
	Games []Game `json:"games"`
}

type Game struct {
	ID             int    `json:"id"`
	GameID         string `json:"gameId"`
	Title          string `json:"title"`
	ReviewId       string `json:"reviewId"`
	ReviewTopImage string `json:"reviewTopImage"`
}

func (gi *GameInfo) FindID(id int) string {
	for _, g := range gi.Games {
		if g.ID == id {
			return g.GameID
		}
	}
	return ""
}

func (gi *GameInfo) GetGameByID(id int) *Game {
	for _, g := range gi.Games {
		if g.ID == id {
			return &g
		}
	}
	return nil
}

func (gi *GameInfo) GetGameByGameID(gameId string) *Game {
	for _, g := range gi.Games {
		if g.GameID == gameId {
			return &g
		}
	}
	return nil
}
