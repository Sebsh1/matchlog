//go:generate mockgen --source=service.go -destination=service_mock.go -package=rating
package rating

import (
	"context"

	"github.com/pkg/errors"
)

type Service interface {
	GetTopXAmongUserIDsByRating(ctx context.Context, topX int, userIDs []uint) (topXUserIDs []uint, ratings []int, err error)
	UpdateRatings(ctx context.Context, draw bool, winningUserIDs, losingUserIDs []uint) error
}

type ServiceImpl struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

func (s *ServiceImpl) GetTopXAmongUserIDsByRating(ctx context.Context, topX int, userIDs []uint) ([]uint, []int, error) {
	userIDs, ratings, err := s.repo.GetTopXAmongUserIDsByRating(ctx, topX, userIDs)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get top %d user ids by rating", topX)
	}

	return userIDs, ratings, nil
}

func (s *ServiceImpl) UpdateRatings(ctx context.Context, draw bool, winningUserIDs, losingUserIDs []uint) error {
	var updatedRatings []Rating

	winnerRatings, err := s.repo.GetRatingsByUserIDs(ctx, winningUserIDs)
	if err != nil {
		return errors.Wrap(err, "failed to get winning users")
	}

	loserRatings, err := s.repo.GetRatingsByUserIDs(ctx, losingUserIDs)
	if err != nil {
		return errors.Wrap(err, "failed to get losing users")
	}

	winnerAverageRating, winnerAverageDeviation := s.getAverageRatingAndDeviation(winnerRatings)
	loserAverageRating, loserAverageDeviation := s.getAverageRatingAndDeviation(loserRatings)

	var winnerResult, loserResult MatchResult

	if draw {
		winnerResult = MatchResult{
			OpponentRating:    loserAverageRating,
			OpponentDeviation: loserAverageDeviation,
			Result:            resultDraw,
		}

		loserResult = MatchResult{
			OpponentRating:    winnerAverageRating,
			OpponentDeviation: winnerAverageDeviation,
			Result:            resultDraw,
		}
	} else {
		winnerResult = MatchResult{
			OpponentRating:    loserAverageRating,
			OpponentDeviation: loserAverageDeviation,
			Result:            resultWin,
		}

		loserResult = MatchResult{
			OpponentRating:    winnerAverageRating,
			OpponentDeviation: winnerAverageDeviation,
			Result:            resultLoss,
		}
	}

	for _, winnerRating := range winnerRatings {
		updatedRating := ApplyActiveRatingPeriod(winnerRating, []MatchResult{winnerResult, loserResult})
		updatedRatings = append(updatedRatings, updatedRating)
	}

	for _, loserRating := range loserRatings {
		updatedRating := ApplyActiveRatingPeriod(loserRating, []MatchResult{loserResult, winnerResult})
		updatedRatings = append(updatedRatings, updatedRating)
	}

	if err := s.repo.UpdateRatings(ctx, updatedRatings); err != nil {
		return errors.Wrap(err, "failed to update ratings")
	}

	return nil
}

func (s *ServiceImpl) getAverageRatingAndDeviation(ratings []Rating) (float64, float64) {
	var totalRating, totalDeviation float64

	for _, rating := range ratings {
		totalRating += rating.Value
		totalDeviation += rating.Deviation
	}

	return totalRating / float64(len(ratings)), totalDeviation / float64(len(ratings))
}
