package subscriptions

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/vechain/thor/block"
	"github.com/vechain/thor/thor"
	"github.com/vechain/thor/tx"
)

type Block struct {
	Number       uint32         `json:"number"`
	ID           thor.Bytes32   `json:"id"`
	Size         uint32         `json:"size"`
	ParentID     thor.Bytes32   `json:"parentID"`
	Timestamp    uint64         `json:"timestamp"`
	GasLimit     uint64         `json:"gasLimit"`
	Beneficiary  thor.Address   `json:"beneficiary"`
	GasUsed      uint64         `json:"gasUsed"`
	TotalScore   uint64         `json:"totalScore"`
	TxsRoot      thor.Bytes32   `json:"txsRoot"`
	StateRoot    thor.Bytes32   `json:"stateRoot"`
	ReceiptsRoot thor.Bytes32   `json:"receiptsRoot"`
	Signer       thor.Address   `json:"signer"`
	Transactions []thor.Bytes32 `json:"transactions"`
	Removed      bool           `json:"removed"`
}

func convertBlock(b *block.Block, removed bool) (*Block, error) {
	if b == nil {
		return nil, nil
	}
	signer, err := b.Header().Signer()
	if err != nil {
		return nil, err
	}
	txs := b.Transactions()
	txIds := make([]thor.Bytes32, len(txs))
	for i, tx := range txs {
		txIds[i] = tx.ID()
	}

	header := b.Header()
	return &Block{
		Number:       header.Number(),
		ID:           header.ID(),
		ParentID:     header.ParentID(),
		Timestamp:    header.Timestamp(),
		TotalScore:   header.TotalScore(),
		GasLimit:     header.GasLimit(),
		GasUsed:      header.GasUsed(),
		Beneficiary:  header.Beneficiary(),
		Signer:       signer,
		Size:         uint32(b.Size()),
		StateRoot:    header.StateRoot(),
		ReceiptsRoot: header.ReceiptsRoot(),
		TxsRoot:      header.TxsRoot(),
		Transactions: txIds,
		Removed:      removed,
	}, nil
}

type Event struct {
	BlockID     thor.Bytes32
	BlockNumber uint32
	BlockTime   uint64
	TxID        thor.Bytes32
	TxOrigin    thor.Address //contract caller
	Address     thor.Address // always a contract address
	Topics      []thor.Bytes32
	Data        []byte
}

func newEvent(header *block.Header, tx *tx.Transaction, event *tx.Event) (*Event, error) {
	origin, err := tx.Signer()
	if err != nil {
		return nil, err
	}

	return &Event{
		header.ID(),
		header.Number(),
		header.Timestamp(),
		tx.ID(),
		origin,
		event.Address,
		event.Topics,
		event.Data,
	}, nil
}

type LogMeta struct {
	BlockID        thor.Bytes32 `json:"blockID"`
	BlockNumber    uint32       `json:"blockNumber"`
	BlockTimestamp uint64       `json:"blockTimestamp"`
	TxID           thor.Bytes32 `json:"txID"`
	TxOrigin       thor.Address `json:"txOrigin"`
}

// FilteredEvent only comes from one contract
type FilteredEvent struct {
	Address thor.Address   `json:"address"`
	Topics  []thor.Bytes32 `json:"topics"`
	Data    string         `json:"data"`
	Meta    LogMeta        `json:"meta"`
	Removed bool           `json:"removed"`
}

//convert a logdb.Event into a json format Event
func convertEvent(event *Event, removed bool) *FilteredEvent {
	return &FilteredEvent{
		Address: event.Address,
		Data:    hexutil.Encode(event.Data),
		Meta: LogMeta{
			BlockID:        event.BlockID,
			BlockNumber:    event.BlockNumber,
			BlockTimestamp: event.BlockTime,
			TxID:           event.TxID,
			TxOrigin:       event.TxOrigin,
		},
		Topics:  event.Topics,
		Removed: removed,
	}
}

// EventFilter contains options for contract event filtering.
type EventFilter struct {
	FromBlock thor.Bytes32  // beginning of the queried range, nil means best block
	Address   *thor.Address // restricts matches to events created by specific contracts

	Topic0 *thor.Bytes32
	Topic1 *thor.Bytes32
	Topic2 *thor.Bytes32
	Topic3 *thor.Bytes32
	Topic4 *thor.Bytes32
}

func (ef *EventFilter) match(event *tx.Event) bool {
	if (ef.Address != nil) && (*ef.Address != event.Address) {
		return false
	}

	matchTopic := func(topic *thor.Bytes32, index int) bool {
		if topic != nil {
			if len(event.Topics) <= index {
				return false
			}

			if *topic != event.Topics[index] {
				return false
			}
		}
		return true
	}

	return matchTopic(ef.Topic0, 0) &&
		matchTopic(ef.Topic1, 1) &&
		matchTopic(ef.Topic2, 2) &&
		matchTopic(ef.Topic3, 3) &&
		matchTopic(ef.Topic4, 4)
}

// TransferFilter contains options for contract transfer filtering.
type TransferFilter struct {
	TxOrigin  *thor.Address // who send transaction
	Sender    *thor.Address // who transferred tokens
	Recipient *thor.Address // who recieved tokens
}
