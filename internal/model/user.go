package model

import "time"

type User struct {
	ID             int64     `db:"id"`
	Username       string    `db:"username"`
	ChatID         int64     `db:"chat_id"`
	Created_at     time.Time `db:"created_at"`
	Approved       bool      `db:"approved"`
	WelcomeMsgSent time.Time `db:"welcome_msg_sent"`
}
