package expay

type (
	// DB is an abstraction of persistent storage
	DB interface {
		Create(v interface{}) (id string, err error)
		Get(id string, v interface{}) error
		Delete(id string) error
		Update(id string, v interface{}) error
		List() (Iter, error)

		// TODO: implement cursor-based pagination
		Paginate(lastCursor string, limit int) (Iter, error)
	}
	// Iter is used to iterate through a list of values
	Iter interface {
		Next() bool
		Scan(v interface{}) (id string, err error)
		Close() error
	}
)

// Payment represents a payment resource
type Payment struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	Version        int               `json:"version"`
	OrganisationID string            `json:"organisation_id"`
	Attributes     PaymentAttributes `json:"attributes"`
}

// PaymentAttributes contains properties of a payment
type PaymentAttributes struct {
	Amount               string             `json:"amount"`
	BeneficiaryParty     BeneficiaryParty   `json:"beneficiary_party"`
	ChargesInformation   ChargesInformation `json:"charges_information"`
	Currency             string             `json:"currency"`
	DebtorParty          DebtorParty        `json:"debtor_party"`
	EndToEndReference    string             `json:"end_to_end_reference"`
	Fx                   Fx                 `json:"fx"`
	NumericReference     string             `json:"numeric_reference"`
	PaymentID            string             `json:"payment_id"`
	PaymentPurpose       string             `json:"payment_purpose"`
	PaymentScheme        string             `json:"payment_scheme"`
	PaymentType          string             `json:"payment_type"`
	ProcessingDate       string             `json:"processing_date"`
	Reference            string             `json:"reference"`
	SchemePaymentSubType string             `json:"scheme_payment_sub_type"`
	SchemePaymentType    string             `json:"scheme_payment_type"`
	SponsorParty         SponsorParty       `json:"sponsor_party"`
}

// BeneficiaryParty represents the beneficiary party of a payment
type BeneficiaryParty struct {
	AccountName       string `json:"account_name"`
	AccountNumber     string `json:"account_number"`
	AccountNumberCode string `json:"account_number_code"`
	AccountType       int    `json:"account_type"`
	Address           string `json:"address"`
	BankID            string `json:"bank_id"`
	BankIDCode        string `json:"bank_id_code"`
	Name              string `json:"name"`
}

// ChargesInformation contains changes information of a payment
type ChargesInformation struct {
	BearerCode              string   `json:"bearer_code"`
	SenderCharges           []Charge `json:"sender_charges"`
	ReceiverChargesAmount   string   `json:"receiver_charges_amount"`
	ReceiverChargesCurrency string   `json:"receiver_charges_currency"`
}

// Charge contains the amount and currency of a change
type Charge struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// DebtorParty represents the debtor party of the payment
type DebtorParty struct {
	AccountName       string `json:"account_name"`
	AccountNumber     string `json:"account_number"`
	AccountNumberCode string `json:"account_number_code"`
	Address           string `json:"address"`
	BankID            string `json:"bank_id"`
	BankIDCode        string `json:"bank_id_code"`
	Name              string `json:"name"`
}

// Fx of a payment
type Fx struct {
	ContractReference string `json:"contract_reference"`
	ExchangeRate      string `json:"exchange_rate"`
	OriginalAmount    string `json:"original_amount"`
	OriginalCurrency  string `json:"original_currency"`
}

// SponsorParty represents the sponsor party of a payment
type SponsorParty struct {
	AccountNumber string `json:"account_number"`
	BankID        string `json:"bank_id"`
	BankIDCode    string `json:"bank_id_code"`
}

// Links of a payment response
type Links struct {
	// self link
	Self string `json:"self,omitempty"`
}

// PaymentResponse is an envelope for a payment response
type PaymentResponse struct {
	// an array of payment
	Data []Payment `json:"data,omitempty"`
	// response links
	Links *Links `json:"links,omitempty"`
}
