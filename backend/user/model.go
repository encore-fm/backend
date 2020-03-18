package user

type Model struct {
	Username string  `bson:"username"`
	Score    float64 `bson:"score"`
}

func New(username string) *Model {
	model := &Model{
		Username: username,
		Score:    0,
	}
	return model
}
