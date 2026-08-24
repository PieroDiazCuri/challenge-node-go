package models

type MatrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

type QRResponse struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

type StatisticsRequest struct {
	Q        [][]float64 `json:"q"`
	R        [][]float64 `json:"r"`
	Original [][]float64 `json:"original"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
