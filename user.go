package main

type user struct {
	ID int `json:"id"`
	Email string `json:"email"`
	Password []byte `json:"password"`
}

type displayUser struct {
	ID int
	Email string
}

func (usr *user) omitPassword() displayUser{
	return displayUser{
		usr.ID,
		usr.Email,
	}
}