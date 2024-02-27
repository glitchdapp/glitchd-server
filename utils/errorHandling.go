package utils

type AccountExists struct{}

func (m *AccountExists) AccountExistsError() string {
	return "Account already exists"
}
