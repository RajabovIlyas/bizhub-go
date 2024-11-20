package models

type Response[T any] struct {
	IsSuccess bool `json:"isSuccess"`
	Result    T    `json:"result"`
	Error     any  `json:"error"`
}

func ErrorResponse(err any) Response[any] {
	return Response[any]{
		IsSuccess: false,
		Result:    nil,
		Error:     err,
	}
}
