package models

type GenericResponseModel struct {
	BError bool
	SError string
	SUrl   string
}

func (r *GenericResponseModel) Err() bool {
	return r.BError
}

func (r *GenericResponseModel) SErr() string {
	return r.SError
}

func (r *GenericResponseModel) Url() string {
	return r.SUrl
}

func (r *GenericResponseModel) SetURL(url string) {
	r.SUrl = url
}
