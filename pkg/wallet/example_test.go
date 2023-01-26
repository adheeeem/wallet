package wallet

import (
	"github.com/adheeeem/wallet/pkg/types"
	"reflect"
	"testing"
)

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
	svc := &Service{}
	account, err := svc.RegisterAccount(types.Phone("+992985570302"))

	if err != nil {
		t.Errorf("Reject(): can't register account %v", err)
		return
	}

	err = svc.Deposit(account.ID, 50_000_00)
	if err != nil {
		t.Errorf("Reject(): can't deposit account, error = %v", err)
		return
	}
	payment, err := svc.Pay(account.ID, 2000_00, "food")

	if err != nil {
		t.Errorf("Reject(): can't create payment, error = %v", err)
		return
	}
	err = svc.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
	}
	got, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, payment) {
		t.Errorf("FindPaymentByID() wrong payment, error = %v", err)
		return
	}
}
