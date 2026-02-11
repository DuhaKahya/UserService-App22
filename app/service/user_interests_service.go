package service

import (
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
)

type userInterestsService struct {
	repo *repository.UserInterestsRepository
}

func NewUserInterestsService(repo *repository.UserInterestsRepository) interfaces.UserInterestsService {
	return &userInterestsService{repo: repo}
}

func (s *userInterestsService) GetForUser(email string) ([]interfaces.UserInterestResponseItem, error) {
	all, err := s.repo.ListAllInterests()
	if err != nil {
		return nil, err
	}

	userRows, err := s.repo.GetUserInterests(email)
	if err != nil {
		return nil, err
	}

	values := map[uint]bool{}
	for _, r := range userRows {
		values[r.InterestID] = r.Value
	}

	resp := make([]interfaces.UserInterestResponseItem, 0, len(all))
	for _, i := range all {
		resp = append(resp, interfaces.UserInterestResponseItem{
			ID:    i.ID,
			Key:   i.Key,
			Value: values[i.ID],
		})
	}

	return resp, nil
}

func (s *userInterestsService) UpdateForUser(email string, input interfaces.UserInterestsUpdateInput) ([]interfaces.UserInterestResponseItem, error) {
	for _, item := range input.Interests {
		row := &models.UserInterest{
			UserEmail:  email,
			InterestID: item.ID,
			Value:      item.Value,
		}
		if err := s.repo.UpsertUserInterest(row); err != nil {
			return nil, err
		}
	}

	return s.GetForUser(email)
}
