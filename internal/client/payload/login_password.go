package payload

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
