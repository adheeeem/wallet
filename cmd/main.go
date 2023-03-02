package main

import (
	"github.com/adheeeem/wallet/pkg/types"
	"github.com/adheeeem/wallet/pkg/wallet"
	"log"
	"os"
)

func main() {
	s := wallet.Service{}
	_, err := s.RegisterAccount("+992985570302")
	_, err = s.RegisterAccount("+992981111111")
	err = s.Deposit(1, 200_00)
	err = s.Deposit(2, 200_00)
	pay, err := s.Pay(1, types.Money(100_00), "grocery")
	_, err = s.Pay(1, types.Money(1_00), "course")
	_, err = s.Pay(1, types.Money(1), "1")
	_, err = s.Pay(1, types.Money(1_0), "2")
	_, err = s.Pay(1, types.Money(1_0), "4")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.FavoritePayment(pay.ID, "shop")
	if err != nil {
		log.Print(err)
		return
	}
	wd, _ := os.Getwd()
	//// s.Export(wd + "/files")
	//s.Import(wd + "/files")
	temp, _ := s.ExportAccountHistory(1)
	_ = s.HistoryToFiles(temp, wd+"/history", 2)

}
