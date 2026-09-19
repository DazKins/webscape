package gameevent

const EventIdBankResolved = "bank:resolved"

type BankResolvedPayload struct {
	Success bool
	Text    string
}
