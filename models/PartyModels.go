package models

type Data struct {
	Party  string `json:"party"`
	Ballot struct {
		Address string `json:"address"`
		Votes   string `json:"votes"`
	} `json:"ballot"`
}
type Ballot struct {
	Signature string `json:"signature"`
	Data      Data   `json:"data"`
}

type Config struct {
	Strategy string `json:"strategy"`
	Nvotes   int32  `json:"nvotes"`
}

type Receipt struct {
	Account string      `json:"account"`
	Amount  interface{} `json:"amount"` // BigNumbers
	Token   string      `json:"token"`
	Txn     string      `json:"txn"`
}
type Party struct {
	ID           string    `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Config       Config    `json:"config"`
	Receipts     []Receipt `json:"receipts"` // Yo
	Participants []string  `json:"participants"`
	Candidates   []string  `json:"candidates"`
	Ballots      []Ballot  `json:"ballots"`
}
