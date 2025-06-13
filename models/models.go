package models

import "encoding/json"

// User 用户模型
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Menu 菜品模型
type Menu struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EnergyCost  int      `json:"energy_cost"`
	ImageURLs   []string `json:"image_urls"`
}

// 自定义 UnmarshalJSON 处理 image_urls
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
	// 解析 image_urls，如果为空则设为 []
	if len(aux.ImageURLs) > 0 {
		if err := json.Unmarshal(aux.ImageURLs, &m.ImageURLs); err != nil {
			return err
		}
	} else {
		m.ImageURLs = []string{}
	}
	return nil
}

// 自定义 MarshalJSON 确保 image_urls 序列化为数组
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

// Party 聚会模型
type Party struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Password   string `json:"password"`
	EnergyLeft int    `json:"energy_left"`
	IsActive   bool   `json:"is_active"`
}

// Order 订单模型
type Order struct {
	ID      int `json:"id"`
	PartyID int `json:"party_id"`
	UserID  int `json:"user_id"`
	MenuID  int `json:"menu_id"`
}
