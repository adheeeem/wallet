package wallet

import "testing"

func TestService_FindAccountByID_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+992985570302")
	svc.RegisterAccount("+992900000000")

	_, err := svc.FindAccountByID(2)
	if err != nil {
		t.Errorf("FindAccountByID, error: %v", ErrAccountNotFound)
	}
}
