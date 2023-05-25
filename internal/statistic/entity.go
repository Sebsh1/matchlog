package statistic

import "time"

type MatchResult int

const (
	ResultWin MatchResult = iota
	ResultLoss
	ResultDraw
)

type Statistic struct {
	ID uint

	Wins          int
	Losses        int
	Draws         int
	MatchesPlayed int

	WinLossRatio float64

	WinStreak  int
	LossStreak int

	CreatedAt time.Time
}
