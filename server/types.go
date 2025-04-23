package main

import "github.com/google/uuid"

type errorResponse struct {
	Error string `json:"error"`
}

type messageResponse struct {
	Message string `json:"message"`
}

type updateWinnerAndLoserParams struct {
	WinnerID uuid.UUID `json:"winner_id"`
	LoserID uuid.UUID `json:"loser_id"` 
}
