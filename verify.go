package expay

// Verify verifies the payment info is in valid format
func (p *Payment) Verify() error {
	if p.Attributes.Amount == "" {
		return ErrInvalidPayment
	}
	return nil
}
