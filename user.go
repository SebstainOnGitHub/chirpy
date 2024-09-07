package main

type user struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type jsonUser struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	//difference is password (string)
	Password string `json:"password"`
}

type displayUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (usr *user) omitPassword() displayUser {
	return displayUser{
		usr.ID,
		usr.Email,
	}
}
