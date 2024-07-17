package utils

import (
	"log"
	"os"
)

func GetCiprLogger() *log.Logger {
	logger := log.New(os.Stdout, "cipr: ", log.LstdFlags)

	return logger
}
