package user_entity

import (

)

type User struct {
    ID      int
}

func NewUser(id int) *User {
	return &User{
		id,
	}
}