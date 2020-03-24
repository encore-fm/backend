package user

type Voter struct {
	Username string  `json:"username" bson:"username"`
	Score    float64 `json:"score" bson:"score"`
}

type Voters []Voter

func NewVoters() Voters {
	return []Voter{}
}

func (v Voters) Find(username string) (int, bool) {
	for i, voter := range v {
		if voter.Username == username {
			return i, true
		}
	}
	return -1, false
}

func (v Voters) Add(username string, score float64) (Voters, bool) {
	_, ok := v.Find(username)
	if ok {
		return v, false
	}
	newVoter := Voter{
		Username: username,
		Score: score,
	}
	voters := []Voter(v)
	voters = append(voters, newVoter)
	return voters, true
}

func (v Voters) Remove(username string) (Voters, bool) {
	i, ok := v.Find(username)
	if !ok {
		return v, false
	}
	voters := []Voter(v)
	voters = append(voters[:i], voters[i+1:]...)
	return voters, true
}

