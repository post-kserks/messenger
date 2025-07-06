package models

// User представляет пользователя системы
type User struct {
    ID       int    `db:"id" json:"id"`
    Username string `db:"username" json:"username"`
    Email    string `db:"email" json:"email"`
    Password string `db:"password" json:"-"` // хеш пароля
}
