package wallet

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/adheeeem/wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

type Progress struct {
	Part   int
	Result types.Money
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	// зачисление средств пока не рассматриваем как платёж
	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}

	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	return s.Pay(payment.AccountID, payment.Amount, payment.Category)
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Amount:    payment.Amount,
		Name:      name,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}

	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}
	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err == nil {
		return nil, err
	}
	return payment, nil
}

func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return err
	}

	for _, account := range s.accounts {
		id := strconv.FormatInt(account.ID, 10)
		phone := account.Phone
		balance := strconv.FormatInt(int64(account.Balance), 10)
		_, err = file.Write([]byte(id + ";" + string(phone) + ";" + balance + "|"))
		if err != nil {
			log.Print(err)
			return err
		}
	}

	return nil
}

func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Print(err)
		}
	}()

	content := make([]byte, 0)
	buf := make([]byte, 4)

	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		content = append(content, buf[:read]...)
	}
	data := string(content)

	splitData := strings.Split(data, "|")
	for _, d := range splitData {
		if d == "" {
			break
		}
		user := strings.Split(d, ";")
		id, _ := strconv.ParseInt(user[0], 10, 64)
		balance, _ := strconv.ParseInt(user[2], 10, 64)
		account := &types.Account{
			ID:      id,
			Phone:   types.Phone(user[1]),
			Balance: types.Money(balance),
		}
		s.accounts = append(s.accounts, account)
	}
	return nil
}

func (s *Service) Export(dir string) error {
	if len(s.accounts) > 0 {
		acc, err := os.Create(dir + "/accounts.dump")
		if err != nil {
			log.Print(err)
			return err
		}
		for _, account := range s.accounts {
			id := strconv.FormatInt(account.ID, 10)
			phone := account.Phone
			balance := strconv.FormatInt(int64(account.Balance), 10)
			_, err = acc.Write([]byte(id + ";" + string(phone) + ";" + balance + "\n"))
			if err != nil {
				log.Print(err)
				return err
			}
		}

	}
	if len(s.payments) > 0 {
		pay, err := os.Create(dir + "/payments.dump")
		if err != nil {
			log.Print(err)
			return err
		}
		for _, payment := range s.payments {
			_, err = pay.Write([]byte(payment.ID + ";" + strconv.Itoa(int(payment.Amount)) + ";" + string(payment.Category) + ";" + string(payment.Status) + ";" + strconv.Itoa(int(payment.AccountID)) + "\n"))
			if err != nil {
				log.Print(err)
				return err
			}
		}
	}
	if len(s.favorites) > 0 {
		fav, err := os.Create(dir + "/favorites.dump")
		if err != nil {
			log.Print(err)
			return err
		}
		for _, favorite := range s.favorites {
			_, err = fav.Write([]byte(favorite.ID + ";" + strconv.Itoa(int(favorite.Amount)) + ";" + string(favorite.Category) + ";" + favorite.Name + ";" + strconv.Itoa(int(favorite.AccountID)) + "\n"))
			if err != nil {
				log.Print(err)
				return err
			}
		}
	}
	return nil
}

func (s *Service) Import(dir string) error {
	acc, err := os.Open(dir + "/accounts.dump")
	if err != nil {
		log.Print(err)
		return err
	}
	pay, err := os.Open(dir + "/payments.dump")
	if err != nil {
		log.Print(err)
		return err
	}
	fav, err := os.Open(dir + "/favorites.dump")
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := acc.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()
	reader := bufio.NewReader(acc)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			log.Print(line)
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		data := strings.Split(line, ";")
		id, _ := strconv.ParseInt(data[0], 10, 64)
		balance, _ := strconv.Atoi(strings.Trim(data[2], "\n"))
		phone := types.Phone(data[1])
		account := &types.Account{
			ID:      id,
			Phone:   phone,
			Balance: types.Money(balance),
		}
		s.accounts = append(s.accounts, account)
	}
	reader = bufio.NewReader(pay)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			log.Print(line)
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		data := strings.Split(line, ";")
		amount, _ := strconv.Atoi(strings.Trim(data[1], "\n"))
		accID, _ := strconv.ParseInt(data[4], 10, 64)
		payment := &types.Payment{
			ID:        data[0],
			Amount:    types.Money(amount),
			Category:  types.PaymentCategory(data[2]),
			Status:    types.PaymentStatus(data[3]),
			AccountID: accID,
		}
		s.payments = append(s.payments, payment)
	}
	reader = bufio.NewReader(fav)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			log.Print(line)
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		data := strings.Split(line, ";")
		amount, _ := strconv.Atoi(strings.Trim(data[1], "\n"))
		accID, _ := strconv.ParseInt(data[4], 10, 64)
		favorite := &types.Favorite{
			ID:        data[0],
			Amount:    types.Money(amount),
			Category:  types.PaymentCategory(data[2]),
			Name:      data[3],
			AccountID: accID,
		}
		s.favorites = append(s.favorites, favorite)
	}
	return nil
}

func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	var pays []types.Payment
	for _, payment := range s.payments {
		if payment.AccountID == accountID {
			pays = append(pays, *payment)
		}
	}
	return pays, nil
}

func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	filesCnt := math.Ceil(float64(len(payments)) / float64(records))
	ind := 0
	if len(payments) <= records {
		pay, err := os.Create(dir + "/payments.dump")
		if err != nil {
			log.Print(err)
			return err
		}
		for j := 0; j < len(payments); j++ {
			id := payments[j].ID
			amount := payments[j].Amount
			category := payments[j].Category
			status := payments[j].Status
			accID := payments[j].AccountID
			_, err = pay.Write([]byte(id + ";" + strconv.Itoa(int(amount)) + ";" + string(category) + ";" + string(status) + ";" + strconv.Itoa(int(accID)) + "\n"))
			if err != nil {
				log.Print(err)
				return err
			}
		}
		return nil
	}
	for i := 1; i <= int(filesCnt); i++ {
		path := fmt.Sprintf(dir+"/payments%d.dump", i)
		pay, err := os.Create(path)
		if err != nil {
			log.Print(err)
			return err
		}
		for j := 0; j < records; j++ {
			id := payments[ind].ID
			amount := payments[ind].Amount
			category := payments[ind].Category
			status := payments[ind].Status
			accID := payments[ind].AccountID
			_, err = pay.Write([]byte(id + ";" + strconv.Itoa(int(amount)) + ";" + string(category) + ";" + string(status) + ";" + strconv.Itoa(int(accID)) + "\n"))
			if err != nil {
				log.Print(err)
				return err
			}
			if ind == len(payments)-1 {
				break
			}
			ind++
		}
	}
	return nil
}

func (s *Service) SumPayments(goroutines int) types.Money {
	mu := sync.Mutex{}
	sum := types.Money(0)
	wg := sync.WaitGroup{}
	for i := 0; i < len(s.payments); i += goroutines {
		j := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			if j+goroutines <= len(s.payments) {
				val := types.Money(0)
				for _, payment := range s.payments[j : j+goroutines] {
					val += payment.Amount
				}
				mu.Lock()
				defer mu.Unlock()
				sum += val
			} else {
				val := types.Money(0)
				for _, payment := range s.payments[j:] {
					val += payment.Amount
				}
				mu.Lock()
				defer mu.Unlock()
				sum += val
			}
		}()
	}
	wg.Wait()

	return sum
}

func (s *Service) FilterPayments(accountID int64, goroutines int) ([]types.Payment, error) {
	wg := sync.WaitGroup{}

	mu := sync.Mutex{}
	var answer []types.Payment
	for i := 0; i < len(s.payments); i += goroutines {
		j := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			if j+goroutines <= len(s.payments) {
				var pays []types.Payment
				for _, payment := range s.payments[j : j+goroutines] {
					if payment.AccountID == accountID {
						pays = append(pays, *payment)
					}
				}
				mu.Lock()
				defer mu.Unlock()
				answer = append(answer, pays...)
			} else {
				var pays []types.Payment
				for _, payment := range s.payments[j:] {
					if payment.AccountID == accountID {
						pays = append(pays, *payment)
					}
				}
				mu.Lock()
				defer mu.Unlock()
				answer = append(answer, pays...)
			}
		}()
	}
	wg.Wait()

	return answer, nil
}

func (s *Service) FilterPaymentByFn(filter func(payment types.Payment) bool, goroutines int) ([]types.Payment, error) {
	wg := sync.WaitGroup{}

	mu := sync.Mutex{}
	var answer []types.Payment
	for i := 0; i < len(s.payments); i += goroutines {
		j := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			if j+goroutines <= len(s.payments) {
				var pays []types.Payment
				for _, payment := range s.payments[j : j+goroutines] {
					if filter(*payment) {
						pays = append(pays, *payment)
					}
				}
				mu.Lock()
				defer mu.Unlock()
				answer = append(answer, pays...)
			} else {
				var pays []types.Payment
				for _, payment := range s.payments[j:] {
					if filter(*payment) {
						pays = append(pays, *payment)
					}
				}
				mu.Lock()
				defer mu.Unlock()
				answer = append(answer, pays...)
			}
		}()
	}
	wg.Wait()

	return answer, nil
}

func (s *Service) SumPaymentsWithProgress() <-chan Progress {
	parts := math.Ceil(float64(len(s.payments) / 100_000))
	size := 100_000
	ch := make(chan Progress)
	for i := 0; i < int(parts); i++ {
		j := i
		go func(ch chan<- Progress, data []*types.Payment) {
			defer close(ch)
			sum := types.Money(0)
			for _, v := range data {
				sum += v.Amount
			}
			ch <- Progress{Part: j, Result: sum}
		}(ch, s.payments[i*size:(i+1)*size])
	}
	return ch
}
