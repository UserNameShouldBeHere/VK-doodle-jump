package domain

type TaskAdminData struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Reward      int    `json:"reward"`
	Token       string `json:"token"`
}
