package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVoters(t *testing.T) {
	expected := Voters([]Voter{})
	result := NewVoters()
	assert.Equal(t, expected, result)
}

func TestVoters_Add(t *testing.T) {
	voters := NewVoters()

	voters, ok := voters.Add("username", 123)
	assert.True(t, ok)
	assert.Equal(t, 1, len(voters))

	voters, ok = voters.Add("username", 13)
	assert.False(t, ok)
	assert.Equal(t, 1, len(voters))

	voters, ok = voters.Add("2", 13)
	assert.True(t, ok)
	assert.Equal(t, 2, len(voters))
}

func TestVoters_Find(t *testing.T) {
	voters := NewVoters()

	voters, _ = voters.Add("username", 123)
	i, ok := voters.Find("username")
	assert.Equal(t, 0, i)
	assert.True(t, ok)

	i, ok = voters.Find("not in voters")
	assert.Equal(t, -1, i)
	assert.False(t, ok)

	voters, _ = voters.Add("user_2", 123)
	i, ok = voters.Find("user_2")
	assert.Equal(t, 1, i)
	assert.True(t, ok)
}

func TestVoters_Remove(t *testing.T) {
	voters := NewVoters()

	voters, _ = voters.Add("1", 123)
	voters, _ = voters.Add("2", 123)
	voters, _ = voters.Add("3", 123)
	assert.Equal(t, 3, len(voters))

	voters, ok := voters.Remove("not in voters")
	assert.False(t, ok)

	voters, ok = voters.Remove("2")
	assert.True(t, ok)
	assert.Equal(t, 2, len(voters))

	i, ok := voters.Find("1")
	assert.True(t, ok)
	assert.Equal(t, 0, i)

	i, ok = voters.Find("2")
	assert.False(t, ok)
	assert.Equal(t, -1, i)

	i, ok = voters.Find("3")
	assert.True(t, ok)
	assert.Equal(t, 1, i)
}