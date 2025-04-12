package domain

import "time"

type PromocodeAdminData struct {
	Id             int       `json:"id,omitempty"`
	Name           string    `json:"name"`
	Company        string    `json:"company"`
	LogoLink       string    `json:"photo_link"`
	Description    string    `json:"description"`
	Price          int       `json:"price"`
	Count          int       `json:"count"`
	Code           string    `json:"code"`
	ActivationLink string    `json:"activation_link"`
	ActiveTo       time.Time `json:"active_to"`
}
