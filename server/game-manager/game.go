package gameManager

type Score struct {
	IsAlive bool `json:"isAlive"`
	Points  int  `json:"points"`
}

type Game struct {
	Id               string
	MaxUserCount     int
	CurrentUserCount int
	WinnerPrice      int
	Entry            int
	Users            map[string]bool
	IsStarted        bool
	ScoreBoard       map[string]Score
}

func (game *Game) UpdateScore(userId string) {
	userScore := game.ScoreBoard[userId]
	if userScore.IsAlive {
		userScore.Points += 1
		game.ScoreBoard[userId] = userScore
	}
}

func (game *Game) GameOver(userId string) {
	userScoreCard, exist := game.ScoreBoard[userId]
	if exist && userScoreCard.IsAlive {
		userScoreCard.IsAlive = false
		game.ScoreBoard[userId] = userScoreCard
	}
}
