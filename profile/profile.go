package profile

type Profile struct {
	GUID  string
	Name  string
	Proxy string
	User  User
}

type User struct {
	Name     string
	LastName string
	Birthday string
	Mobile   string
	Email    string
}
