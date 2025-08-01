package sync

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type SyncManager struct {
	trackedStates map[uint64]interface{}
	stateLock     sync.RWMutex
}

func NewSyncManager() *SyncManager {
	return &SyncManager{
		trackedStates: make(map[uint64]interface{}),
	}
}

func (sm *SyncManager) TrackEntity(entityID uint64, initialState interface{}) {
	sm.stateLock.Lock()
	defer sm.stateLock.Unlock()

	sm.trackedStates[entityID] = initialState
	fmt.Printf("[SyncManager] Started tracking entity %d\n", entityID)
}

func (sm *SyncManager) StopTracking(entityID uint64) {
	sm.stateLock.Lock()
	defer sm.stateLock.Unlock()
	delete(sm.trackedStates, entityID)
	fmt.Printf("[SyncManager] Stopped tracking entity %d\n", entityID)
}

func (sm *SyncManager) GenerateSyncMessage(entityID uint64, currentState interface{}, forceFullSync bool) (*SyncMessage, bool, error) {
	sm.stateLock.Lock()
	defer sm.stateLock.Unlock()

	lastKnownState, exists := sm.trackedStates[entityID]

	if forceFullSync || !exists {
		fmt.Printf("[SyncManager] Generating FULL sync for entity %d (forceFullSync=%v, exists=%v)\n", entityID, forceFullSync, exists)

		sm.trackedStates[entityID] = currentState
		return &SyncMessage{
			MsgType:  SyncTypeFull,
			EntityID: entityID,
			State:    currentState,
		}, true, nil
	}

	deltas, err := GenerateDelta(lastKnownState, currentState)
	if err != nil {

		fmt.Printf("[SyncManager] Error generating delta for entity %d: %v. Falling back to FULL sync.\n", entityID, err)
		sm.trackedStates[entityID] = currentState
		return &SyncMessage{
			MsgType:  SyncTypeFull,
			EntityID: entityID,
			State:    currentState,
		}, true, nil

	}

	if len(deltas) == 0 {

		sm.trackedStates[entityID] = currentState
		return nil, false, nil
	}

	sm.trackedStates[entityID] = currentState
	return &SyncMessage{
		MsgType:  SyncTypeDelta,
		EntityID: entityID,
		DeltaSet: deltas,
	}, true, nil
}

func (sm *SyncManager) ApplyReceivedSyncMessage(entityID uint64, currentStatePtr interface{}, msg *SyncMessage) error {
	if msg == nil {
		return errors.New("received nil sync message")
	}
	if msg.EntityID != entityID {
		return fmt.Errorf("sync message entity ID mismatch: expected %d, got %d", entityID, msg.EntityID)
	}

	if msg.MsgType == SyncTypeFull {

		fmt.Printf("[SyncManager] Applying FULL sync for entity %d\n", entityID)

		targetVal := reflect.ValueOf(currentStatePtr)
		if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
			return ErrNotPtr
		}
		targetElem := targetVal.Elem()
		newStateVal := reflect.ValueOf(msg.State)

		if !newStateVal.IsValid() {

			return errors.New("received full sync with nil state")
		}
		if targetElem.Type() != newStateVal.Type() {

			if newStateVal.Kind() == reflect.Interface {
				newStateVal = newStateVal.Elem()
			}
			if targetElem.Type() != newStateVal.Type() {
				return fmt.Errorf("type mismatch for full sync: target=%s, received=%s", targetElem.Type(), newStateVal.Type())
			}
		}
		if !targetElem.CanSet() {
			return errors.New("cannot set target state for full sync")
		}
		targetElem.Set(newStateVal)

	} else if msg.MsgType == SyncTypeDelta {

		fmt.Printf("[SyncManager] Applying DELTA sync for entity %d (%d deltas)\n", entityID, len(msg.DeltaSet))
		if err := ApplyDelta(currentStatePtr, msg.DeltaSet); err != nil {
			return fmt.Errorf("failed to apply delta sync for entity %d: %w", entityID, err)
		}

	} else {
		return fmt.Errorf("unknown sync message type: %d", msg.MsgType)
	}

	return nil
}
