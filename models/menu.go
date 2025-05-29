package models

type Menu struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    EnergyCost  int    `json:"energy_cost"`
}