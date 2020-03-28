package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVoters(t *testing.T) {
	expected := voters(make(map[string]float64))
	result := NewVoters()
	assert.Equal(t, expected, result)
}

func TestVoters_Add(t *testing.T) {
	voters := NewVoters()

	ok := voters.Add("username", 123)
	assert.True(t, ok)
	assert.Equal(t, 1, voters.Size())

	ok = voters.Add("username", 13)
	assert.False(t, ok)
	assert.Equal(t, 1, voters.Size())

	ok = voters.Add("2", 13)
	assert.True(t, ok)
	assert.Equal(t, 2, voters.Size())
}

func TestVoters_Find(t *testing.T) {
	voters := NewVoters()

	_ = voters.Add("username", 123)
	ok := voters.Contains("username")
	assert.True(t, ok)

	ok = voters.Contains("not in voters")
	assert.False(t, ok)

	_ = voters.Add("user_2", 123)
	ok = voters.Contains("user_2")
	assert.True(t, ok)
}

func TestVoters_Remove(t *testing.T) {
	voters := NewVoters()

	 _ = voters.Add("1", 123)
	 _ = voters.Add("2", 123)
	 _ = voters.Add("3", 123)
	assert.Equal(t, 3, voters.Size())

	ok := voters.Remove("not in voters")
	assert.False(t, ok)

	ok = voters.Remove("2")
	assert.True(t, ok)
	assert.Equal(t, 2, voters.Size())

	ok = voters.Contains("1")
	assert.True(t, ok)

	ok = voters.Contains("2")
	assert.False(t, ok)

	ok = voters.Contains("3")
	assert.True(t, ok)
}