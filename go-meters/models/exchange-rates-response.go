package models

type ExchangeRatesResponseModel struct {
	GenericResponseModel
	Rates []Rate `json:"rates"`
}
