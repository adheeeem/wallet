package main

import (
	"github.com/adheeeem/wallet/pkg/wallet"
	"log"
	"os"
)

func main() {
	s := wallet.Service{}
	_, err := s.RegisterAccount("+992985570302")
	_, err = s.RegisterAccount("+992981111111")
	if err != nil {
		log.Print(err)
		return
	}
	wd, err := os.Getwd()
	s.ExportToFile(wd + "/file.txt")
	// s.ImportFromFile(wd + "/file.txt")
}
