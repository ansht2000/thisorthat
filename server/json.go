package main

func returnErrJSON(errMessage string) errorResponse {
	errRes := errorResponse{
		Error: errMessage,
	}
	return errRes
}

func returnMessageJSON(message string) messageResponse {
	messRes := messageResponse{
		Message: message,
	}
	return messRes
}