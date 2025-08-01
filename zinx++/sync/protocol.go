package sync

import "encoding/json"

type SyncMsgType uint8

const (
	SyncTypeFull  SyncMsgType = 1
	SyncTypeDelta SyncMsgType = 2
)

type Delta struct {
	FieldName string      `json:"f"`
	NewValue  interface{} `json:"v"`
}

type SyncMessage struct {
	MsgType     SyncMsgType `json:"mt"`
	EntityID    uint64      `json:"eid"`
	State       interface{} `json:"state,omitempty"`
	DeltaSet    []Delta     `json:"delta,omitempty"`
	BaseVersion uint64      `json:"bv,omitempty"`
}

func MarshalSyncMessage(msg *SyncMessage) ([]byte, error) {

	return json.Marshal(msg)
}

func UnmarshalSyncMessage(data []byte) (*SyncMessage, error) {
	var msg SyncMessage

	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
