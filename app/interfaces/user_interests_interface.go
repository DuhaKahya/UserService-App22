package interfaces

type UserInterestResponseItem struct {
	ID    uint   `json:"id"`
	Key   string `json:"key"`
	Value bool   `json:"value"`
}

type UserInterestItemInput struct {
	ID    uint `json:"id"`
	Value bool `json:"value"`
}

type UserInterestsUpdateInput struct {
	Interests []UserInterestItemInput `json:"interests"`
}

type UserInterestsService interface {
	GetForUser(email string) ([]UserInterestResponseItem, error)
	UpdateForUser(email string, input UserInterestsUpdateInput) ([]UserInterestResponseItem, error)
}
