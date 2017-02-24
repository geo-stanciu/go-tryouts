package models

type LoginResponse struct {
	BError bool
	SError string
	URL    string
}

func (l *LoginResponse) Err() bool {
	return l.BError
}

func (l *LoginResponse) SErr() string {
	return l.SError
}

func (l *LoginResponse) Url() string {
	return l.URL
}

func (l *LoginResponse) SetURL(url string) {
	l.URL = url
}
