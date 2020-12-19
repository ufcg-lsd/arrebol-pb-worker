package errors

import (
	"log"
	"os"
)

type CustomError struct {
	Code int
	Message string
}

func (e *CustomError) String() string{
	return e.Message + "\n Code: " + string(e.Code)
}

func (e *CustomError) panic() {
	log.Println(e.Message)
	os.Exit(e.Code)
}
