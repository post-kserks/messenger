package models

// Chat представляет чат (групповой или личный)
type Chat struct {
    ID      int    `db:"id" json:"id"`
    Name    string `db:"name" json:"name"`
    IsGroup bool   `db:"is_group" json:"is_group"`
}
