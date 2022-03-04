package models

type Data struct {
	Party  string `json:"party"`
	Ballot struct {
		Votes     string `json:"votes"`
		Timestamp string `json:"timestamp"`
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
	Account  string      `json:"account"`
	Amount   interface{} `json:"amount"` // BigNumbers
	Token    string      `json:"token"`
	Txn      string      `json:"txn"`
	Strategy string      `json:"strategy"`
	ChainId  uint16      `json:"chainId"`
}

type Note struct {
	Candidate string `json:"candidate"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type SignedParty struct {
	Data struct {
		Version      string   `json:"version"`
		Name         string   `json:"name"`
		Timestamp    string   `json:"timestamp"`
		Nonce        string   `json:"nonce"`
		Description  string   `json:"description"`
		Participants []string `json:"participants"`
		Candidates   []string `json:"candidates"`
	} `json:"data"`
	Signature string `json:"signature"`
}

type Party struct {
	ID           string      `json:"id,omitempty" bson:"_id,omitempty"`
	Version      string      `json:"version"`
	Name         string      `json:"name"`
	Timestamp    string      `json:"timestamp"`
	Nonce        string      `json:"nonce"`
	Description  string      `json:"description"`
	Config       Config      `json:"config"`
	Receipts     []Receipt   `json:"receipts"`
	Participants []string    `json:"participants"`
	Candidates   []string    `json:"candidates"`
	Ballots      []Ballot    `json:"ballots"`
	Notes        []Note      `json:"notes"`
	IPFS         string      `json:"ipfs"`
	Signed       SignedParty `json:"signed"`
}
