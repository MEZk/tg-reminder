package domain

import (
	"fmt"
	"time"
)

type User struct {
	ID         int64      `db:"id"`
	Name       string     `db:"name"`
	Status     UserStatus `db:"status"`
	CreatedAt  time.Time  `db:"created_at"`
	ModifiedAt time.Time  `db:"modified_at"`
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

func (u User) String() string {
	return fmt.Sprintf("[ID: %d, Name: %s, Status: %s]", u.ID, u.Name, u.Status)
}

var locationMSK, _ = time.LoadLocation("Europe/Moscow")

// TODO: ask user timezone
func MoscowTime(t time.Time) time.Time {
	return t.In(locationMSK)
}
