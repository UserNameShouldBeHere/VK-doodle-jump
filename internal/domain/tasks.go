package domain

type TaskAdminData struct {
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Token       string `json:"token"`
}

type TaskData struct {
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}
