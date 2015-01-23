package braintree

import (
	"math/big"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func offset() *big.Rat {
	return big.NewRat(rand.Int63n(100), 1)
}

func TestTransactionCreateSubmitForSettlementAndVoid(t *testing.T) {
	amount := *big.NewRat(0, 1).Add(big.NewRat(130, 1), offset())

	tx, err := testGateway.Transaction().Create(&Transaction{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCreditCards["visa"].Number,
			ExpirationDate: "05/14",
		},
	})

	t.Log(tx)

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("Received invalid ID on new transaction")
	}
	if tx.Status != "authorized" {
		t.Fatal(tx.Status)
	}

	// Submit for settlement
	ten := *big.NewRat(10, 1)
	tx2, err := testGateway.Transaction().SubmitForSettlement(tx.Id, ten)

	t.Log(tx2)

	if err != nil {
		t.Fatal(err)
	}
	if status := tx2.Status; status != "submitted_for_settlement" {
		t.Fatal(status)
	}
	if amount := tx2.Amount; ten.Cmp(&amount) != 0 {
		t.Fatal(amount)
	}

	// Void
	tx3, err := testGateway.Transaction().Void(tx2.Id)

	t.Log(tx3)

	if err != nil {
		t.Fatal(err)
	}
	if x := tx3.Status; x != "voided" {
		t.Fatal(x)
	}
}

func TestTransactionSearch(t *testing.T) {
	txg := testGateway.Transaction()
	createTx := func(amount big.Rat, customerName string) error {
		_, err := txg.Create(&Transaction{
			Type:   "sale",
			Amount: amount,
			Customer: &Customer{
				FirstName: customerName,
			},
			CreditCard: &CreditCard{
				Number:         testCreditCards["visa"].Number,
				ExpirationDate: "05/14",
			},
		})
		return err
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	name := "Erik-" + ts

	amountOne := *big.NewRat(0, 1).Add(big.NewRat(100, 1), offset())
	if err := createTx(amountOne, name); err != nil {
		t.Fatal(err)
	}

	amountTwo := *big.NewRat(0, 1).Add(big.NewRat(150, 1), offset())
	if err := createTx(amountTwo, "Lionel-"+ts); err != nil {
		t.Fatal(err)
	}

	query := new(SearchQuery)
	f := query.AddTextField("customer-first-name")
	f.Is = name

	result, err := txg.Search(query)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.TotalItems) != 1 {
		t.Fatal(result.Transactions)
	}

	tx := result.Transactions[0]
	if x := tx.Customer.FirstName; x != name {
		t.Log(name)
		t.Fatal(x)
	}
}

// This test will fail unless you set up your Braintree sandbox account correctly. See TESTING.md for details.
func TestTransactionCreateWhenGatewayRejected(t *testing.T) {
	amount := *big.NewRat(2010, 1)

	_, err := testGateway.Transaction().Create(&Transaction{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCreditCards["visa"].Number,
			ExpirationDate: "05/14",
		},
	})
	if err == nil {
		t.Fatal("Did not receive error when creating invalid transaction")
	}
	if err.Error() != "Card Issuer Declined CVV" {
		t.Fatal(err)
	}
}

func TestFindTransaction(t *testing.T) {
	amount := *big.NewRat(0, 1).Add(big.NewRat(110, 1), offset())

	createdTransaction, err := testGateway.Transaction().Create(&Transaction{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCreditCards["mastercard"].Number,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	foundTransaction, err := testGateway.Transaction().Find(createdTransaction.Id)
	if err != nil {
		t.Fatal(err)
	}

	if createdTransaction.Id != foundTransaction.Id {
		t.Fatal("transaction ids do not match")
	}
}

func TestFindNonExistantTransaction(t *testing.T) {
	_, err := testGateway.Transaction().Find("bad_transaction_id")
	if err == nil {
		t.Fatal("Did not receive error when finding an invalid tx ID")
	}
	if err.Error() != "Not Found (404)" {
		t.Fatal(err)
	}
}

func TestAllTransactionFields(t *testing.T) {
	amount := *big.NewRat(0, 1).Add(big.NewRat(100, 1), offset())

	tx := &Transaction{
		Type:    "sale",
		Amount:  amount,
		OrderId: "my_custom_order",
		CreditCard: &CreditCard{
			Number:         testCreditCards["visa"].Number,
			ExpirationDate: "05/14",
			CVV:            "100",
		},
		Customer: &Customer{
			FirstName: "Lionel",
		},
		BillingAddress: &Address{
			StreetAddress: "1 E Main St",
			Locality:      "Chicago",
			Region:        "IL",
			PostalCode:    "60637",
		},
		ShippingAddress: &Address{
			StreetAddress: "1 E Main St",
			Locality:      "Chicago",
			Region:        "IL",
			PostalCode:    "60637",
		},
		Options: &TransactionOptions{
			SubmitForSettlement:              true,
			StoreInVault:                     true,
			AddBillingAddressToPaymentMethod: true,
			StoreShippingAddressInVault:      true,
		},
	}

	tx2, err := testGateway.Transaction().Create(tx)
	if err != nil {
		t.Fatal(err)
	}

	if tx2.Type != tx.Type {
		t.Fail()
	}
	if tx2.Amount.Cmp(&tx.Amount) != 0 {
		t.Fail()
	}
	if tx2.OrderId != tx.OrderId {
		t.Fail()
	}
	if tx2.Customer.FirstName != tx.Customer.FirstName {
		t.Fail()
	}
	if tx2.BillingAddress.StreetAddress != tx.BillingAddress.StreetAddress {
		t.Fail()
	}
	if tx2.BillingAddress.Locality != tx.BillingAddress.Locality {
		t.Fail()
	}
	if tx2.BillingAddress.Region != tx.BillingAddress.Region {
		t.Fail()
	}
	if tx2.BillingAddress.PostalCode != tx.BillingAddress.PostalCode {
		t.Fail()
	}
	if tx2.ShippingAddress.StreetAddress != tx.ShippingAddress.StreetAddress {
		t.Fail()
	}
	if tx2.ShippingAddress.Locality != tx.ShippingAddress.Locality {
		t.Fail()
	}
	if tx2.ShippingAddress.Region != tx.ShippingAddress.Region {
		t.Fail()
	}
	if tx2.ShippingAddress.PostalCode != tx.ShippingAddress.PostalCode {
		t.Fail()
	}
	if tx2.CreditCard.Token == "" {
		t.Fail()
	}
	if tx2.Customer.Id == "" {
		t.Fail()
	}
	if tx2.Status != "submitted_for_settlement" {
		t.Fail()
	}
}

// This test will only pass on Travis. See TESTING.md for more details.
func TestTransactionDisbursementDetails(t *testing.T) {
	txn, err := testGateway.Transaction().Find("dskdmb")
	if err != nil {
		t.Fatal(err)
	}

	if txn.DisbursementDetails.DisbursementDate != "2013-06-27" {
		t.Fail()
	}
	if txn.DisbursementDetails.SettlementAmount != "100.00" {
		t.Fail()
	}
	if txn.DisbursementDetails.SettlementCurrencyIsoCode != "USD" {
		t.Fail()
	}
	if txn.DisbursementDetails.SettlementCurrencyExchangeRate != "1" {
		t.Fail()
	}
	if txn.DisbursementDetails.FundsHeld == "true" {
		t.Fail()
	}
	if txn.DisbursementDetails.Success != "true" {
		t.Fail()
	}
}

func TestTransactionCreateFromPaymentMethodCode(t *testing.T) {
	customer, err := testGateway.Customer().Create(&Customer{
		CreditCard: &CreditCard{
			Number:         testCreditCards["discover"].Number,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if customer.CreditCards.CreditCard[0].Token == "" {
		t.Fatal("invalid token")
	}

	amount := *big.NewRat(0, 1).Add(big.NewRat(120, 1), offset())

	tx, err := testGateway.Transaction().Create(&Transaction{
		Type:               "sale",
		CustomerID:         customer.Id,
		Amount:             amount,
		PaymentMethodToken: customer.CreditCards.CreditCard[0].Token,
	})

	if err != nil {
		t.Fatal(err)
	}
	if tx.Id == "" {
		t.Fatal("invalid tx id")
	}
}

func TestSettleTransaction(t *testing.T) {
	old_environment := testGateway.Environment
	amount := *big.NewRat(0, 1).Add(big.NewRat(130, 1), offset())

	txn, err := testGateway.Transaction().Create(&Transaction{
		Type:   "sale",
		Amount: amount,
		CreditCard: &CreditCard{
			Number:         testCreditCards["visa"].Number,
			ExpirationDate: "05/14",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txn, err = testGateway.Transaction().SubmitForSettlement(txn.Id, txn.Amount)
	if err != nil {
		t.Fatal(err)
	}

	testGateway.Environment = Production

	_, err = testGateway.Transaction().Settle(txn.Id)
	if err.Error() != "Operation not allowed in production environment" {
		t.Log(testGateway.Environment)
		t.Fatal(err)
	}

	testGateway.Environment = old_environment

	txn, err = testGateway.Transaction().Settle(txn.Id)
	if err != nil {
		t.Fatal(err)
	}

	if txn.Status != "settled" {
		t.Fatal(txn.Status)
	}
}
