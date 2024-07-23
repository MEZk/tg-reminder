package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMoscowTime(t *testing.T) {
	t.Parallel()
	date := time.Date(2024, 1, 1, 23, 1, 1, 0, time.UTC)
	assert.Equal(t, time.Date(2024, 1, 2, 2, 1, 1, 0, locationMSK), MoscowTime(date))
}

func TestUser_String(t *testing.T) {
	t.Parallel()
	user := User{
		ID:     1,
		Name:   "Monya Grindstaff",
		Status: UserStatusActive,
	}
	assert.Equal(t, "[ID: 1, Name: Monya Grindstaff, Status: active]", user.String())

	assert.EqualValues(t, "active", UserStatusActive)
	assert.EqualValues(t, "inactive", UserStatusInactive)
}
