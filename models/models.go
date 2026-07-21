package models

import "encoding/json"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

type Menu struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EnergyCost  int      `json:"energy_cost"`
	ImageURLs   []string `json:"image_urls"`
}

func (m *Menu) UnmarshalJSON(data []byte) error {
	type Alias Menu
	aux := &struct {
		*Alias
		ImageURLs json.RawMessage `json:"image_urls"`
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.ImageURLs) > 0 {
		if err := json.Unmarshal(aux.ImageURLs, &m.ImageURLs); err != nil {
			return err
		}
	} else {
		m.ImageURLs = []string{}
	}
	return nil
}

func (m Menu) MarshalJSON() ([]byte, error) {
	type Alias Menu
	return json.Marshal(&struct {
		*Alias
		ImageURLs []string `json:"image_urls"`
	}{
		Alias:     (*Alias)(&m),
		ImageURLs: m.ImageURLs,
	})
}

type Party struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Password   string `json:"-"`
	EnergyLeft int    `json:"energy_left"`
	IsActive   bool   `json:"is_active"`
}

type PartyMember struct {
	ID       int `json:"id"`
	PartyID  int `json:"party_id"`
	UserID   int `json:"user_id"`
}

type Order struct {
	ID      int `json:"id"`
	PartyID int `json:"party_id"`
	UserID  int `json:"user_id"`
	MenuID  int `json:"menu_id"`
}
