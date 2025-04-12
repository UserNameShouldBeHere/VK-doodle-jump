package domain

type ProductAdminData struct {
	Id             int    `json:"id,omitempty"`
	Name           string `json:"name"`
	PhotoLink      string `json:"photo_link"`
	Description    string `json:"description"`
	Price          int    `json:"priсe"`
	Count          int    `json:"count"`
	ActivationLink string `json:"activation_link"`
}
