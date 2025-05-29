package models

type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Password string `json:"password"`
    Role     string `json:"role"`
}

type Menu struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    EnergyCost  int    `json:"energy_cost"`
}

type Party struct {
    ID         int    `json:"id"`
    Name       string `json:"name"`
    Password   string `json:"password"`
    EnergyLeft int    `json:"energy_left"`
    IsActive   bool   `json:"is_active"`
}

type Order struct {
    ID      int `json:"id"`
    PartyID int `json:"party_id"`
    UserID  int `json:"user_id"`
    MenuID  int `json:"menu_id"`
}