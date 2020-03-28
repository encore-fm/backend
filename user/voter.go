package user

type Voters map[string]float64

func NewVoters() Voters {
	return make(map[string]float64)
}

func (v Voters) Contains(username string) bool {
	_, ok := v[username]
	return ok
}

func (v Voters) Get(username string) (float64, bool){
	val, ok := v[username]
	return val, ok
}

func (v Voters) Add(username string, score float64) bool {
	ok := v.Contains(username)
	if ok {
		return false
	}
	v[username] = score
	return true
}

func (v Voters) Remove(username string) bool {
	if ok := v.Contains(username); !ok {
		return false
	}
	delete(v, username)
	return true
}

func (v Voters) Size() int {
	return len(v)
}
