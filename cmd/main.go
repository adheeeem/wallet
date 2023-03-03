package main

import (
	"github.com/adheeeem/wallet/pkg/types"
	"github.com/adheeeem/wallet/pkg/wallet"
	"log"
)

func main() {
	s := wallet.Service{}
	_, err := s.RegisterAccount("+992985570302")
	_, err = s.RegisterAccount("+992981111111")
	err = s.Deposit(1, 1000_00)
	err = s.Deposit(2, 1000_00)
	pay, err := s.Pay(1, types.Money(100_00), "grocery")
	_, err = s.Pay(1, types.Money(1_00), "course")
	_, err = s.Pay(1, types.Money(1), "1")
	_, err = s.Pay(1, types.Money(1_0), "2")
	_, err = s.Pay(1, types.Money(1_0), "4")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.Pay(2, types.Money(1_00), "course")
	_, err = s.FavoritePayment(pay.ID, "shop")
	if err != nil {
		log.Print(err)
		return
	}
	sm := s.SumPayments(5)
	log.Print(sm)

}
