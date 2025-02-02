package models

import "time"

type Vote struct {
	VoteString string     `json:"vote_string"`
	IP         string     `json:"ip"`
	Location   IPLocation `json:"location"`
	Country    string     `json:"country"`
	Year       int        `json:"year"`
	Timestamp  time.Time  `json:"timestamp"`
}
