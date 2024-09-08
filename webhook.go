package main

type webhookBody struct {
	Event string `json:"event"`
	Data webhookData
}

type webhookData struct {
	ID int `json:"user_id"`
}
