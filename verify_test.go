package expay

import (
	"encoding/json"
	"testing"

	"h12.io/expay/testdata"
)

func TestPaymentVerify(t *testing.T) {
	t.Parallel()

	payment := Payment{}
	_ = json.Unmarshal([]byte(testdata.Payment), &payment)
	testcases := []struct {
		name string
		pay  Payment

		wantErr error
	}{
		{
			name:    "invalid payment",
			pay:     Payment{},
			wantErr: ErrInvalidPayment,
		},
		{
			name:    "valid payment",
			pay:     payment,
			wantErr: nil,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.pay.Verify(); err != tc.wantErr {
				t.Fatalf("expect error %v got %v", tc.wantErr, err)
			}
		})
	}
}
