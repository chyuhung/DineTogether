package models

type Order struct {
    ID      int `json:"id"`
    PartyID int `json:"party_id"`
    UserID  int `json:"user_id"`
    MenuID  int `json:"menu_id"`
}