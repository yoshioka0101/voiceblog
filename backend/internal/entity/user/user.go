package user

type User struct {
	ID           int64
	AuthProvider string
	AuthSubject  string
	Email        string
	Name         string
}
