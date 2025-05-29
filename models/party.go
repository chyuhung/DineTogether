package models

type Party struct {
    ID         int    `json:"id"`
    Name       string `json:"name"`
    Password   string `json:"password"`
    EnergyLeft int    `json:"energy_left"`
    IsActive   bool   `json:"is_active"`
}