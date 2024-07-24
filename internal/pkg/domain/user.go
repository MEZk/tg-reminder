package domain

import (
	"fmt"
	"time"
)

// User describes user.
type User struct {
	ID         int64      `db:"id"`
	Name       string     `db:"name"`
	Status     UserStatus `db:"status"`
	CreatedAt  time.Time  `db:"created_at"`
	ModifiedAt time.Time  `db:"modified_at"`
}

// UserStatus is a user status.
type UserStatus string

const (
	// UserStatusActive - user is active and is able to receive reminders.
	UserStatusActive UserStatus = "active"
	// UserStatusInactive  - user is not active and is not able to receive reminders.
	UserStatusInactive UserStatus = "inactive"
)

// String implements [fmt.Stringer].
func (u User) String() string {
	return fmt.Sprintf("[ID: %d, Name: %s, Status: %s]", u.ID, u.Name, u.Status)
}

var locationMSK, _ = time.LoadLocation("Europe/Moscow")

// MoscowTime returns "t" in "Europe/Moscow" location.
// TODO: ask user timezone
func MoscowTime(t time.Time) time.Time {
	return t.In(locationMSK)
}
