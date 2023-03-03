package wallet

import (
	"fmt"
	"github.com/adheeeem/wallet/pkg/types"
	"github.com/google/uuid"
	"log"
	"reflect"
	"testing"
)

type testService struct {
	*Service
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = testAccount{
	phone:   "+992985570302",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func newTestService() *testService {
	return &testService{&Service{}}
}

func (s *Service) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	account, err := s.RegisterAccount(phone)

	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposit account, error = %v", err)
	}
	return account, nil
}

func TestService_FindAccountByID_success(t *testing.T) {
	svc := &Service{}
	svc.RegisterAccount("+992985570302")
	svc.RegisterAccount("+992900000000")

	_, err := svc.FindAccountByID(2)
	if err != nil {
		t.Errorf("FindAccountByID, error: %v", ErrAccountNotFound)
	}
}
func TestService_Reject_success(t *testing.T) {
	svc := newTestService()

	_, payments, err := svc.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	payment := payments[0]
	err = svc.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}

	savedPayment, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject() can't find payment by id, error = %v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't change, payment = %v", savedPayment)
		return
	}
	savedAccount, err := svc.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't change, account = %v", savedAccount)
	}
}
func TestService_FindPaymentByID_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID() wrong payment returned = %v", err)
		return
	}
}
func TestService_FindPaymentByID_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Error("FindPaymentByID() must return error, returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPAymentNotFound, returned = %v", err)
		return
	}
}
func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposit account, error = %v", err)
	}
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}
	return account, payments, nil
}
func TestService_Repeat_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)

	if err != nil {
		t.Errorf("Repeat(): can't add account, error = %v", err)
		return
	}
	payment := payments[0]
	otherPayment, err := s.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Repeat(): can't repeat payment, error = %v", err)
		return
	}

	if payment.ID == otherPayment.ID {
		t.Errorf("Repeat(): two payments have same ids")
		return
	}
}

func TestService_FavoritePayment(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("Repeat(): can't add account, error = %v", err)
		return
	}
	payment := payments[0]
	_, err = s.FavoritePayment(payment.ID, "Alif Course")
	if err != nil {
		t.Errorf("FavoritePayment(): can't create a favorite payment, error = %v", err)
		return
	}
}

func TestService_PayFromFavorite(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("Repeat(): can't add account, error = %v", err)
		return
	}
	payment := payments[0]
	favoritePayment, err := s.FavoritePayment(payment.ID, "Alif Course")
	if err != nil {
		t.Errorf("PayFromFavorite(): can't create a favorite payment, error = %v", err)
		return
	}
	_, err = s.PayFromFavorite(favoritePayment.ID)
	if err != nil {
		t.Errorf("PayFromFavorite(): can't pay from favorite payment, error = %v", err)
		return
	}
}
func BenchmarkService_SumPayments(b *testing.B) {
	s := newTestService()
	_, err := s.RegisterAccount("+992985570302")
	_, err = s.RegisterAccount("+992981111111")

	if err != nil {
		log.Print(err)
	}

	err = s.Deposit(1, 1000_00)
	err = s.Deposit(2, 1000_00)

	if err != nil {
		log.Print(err)
	}

	_, err = s.Pay(1, types.Money(100_00), "grocery")
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

	if err != nil {
		log.Print(err)
	}

	want := types.Money(11121)
	for i := 0; i < b.N; i++ {
		result := s.SumPayments(2)
		if result != want {
			b.Fatalf("invalid result, got %v, want %v", result, want)
		}
	}
}

func BenchmarkService_FilterPayments(b *testing.B) {
	s := newTestService()
	_, err := s.RegisterAccount("+992985570302")
	_, err = s.RegisterAccount("+992981111111")

	if err != nil {
		log.Print(err)
	}

	err = s.Deposit(1, 1000_000_00)
	err = s.Deposit(2, 1000_000_00)

	if err != nil {
		log.Print(err)
	}

	_, err = s.Pay(1, types.Money(100_00), "grocery")

	for i := 0; i < 1000; i++ {
		_, err = s.Pay(1, types.Money(10), "grocery")
	}

	if err != nil {
		log.Print(err)
	}
	for i := 0; i < b.N; i++ {
		_, err := s.FilterPayments(1, 20)
		if err != nil {
			log.Println(err)
		}
	}
}
