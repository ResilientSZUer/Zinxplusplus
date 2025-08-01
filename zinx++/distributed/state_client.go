package distributed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"zinxplusplus/state"
	"zinxplusplus/ziface"
)

type StateClient struct {
	manager ziface.IStateManager
}

func NewStateClient(stateMgr ziface.IStateManager) (*StateClient, error) {
	if stateMgr == nil {
		return nil, errors.New("state manager cannot be nil")
	}
	return &StateClient{
		manager: stateMgr,
	}, nil
}

func (sc *StateClient) GetPlayerState(ctx context.Context, playerID uint64, statePtr interface{}) error {
	stateKey := fmt.Sprintf("player:%d", playerID)

	if getter, ok := sc.manager.(interface {
		GetStateObject(context.Context, string, interface{}) error
	}); ok {
		err := getter.GetStateObject(ctx, stateKey, statePtr)
		if err != nil {

			if errors.Is(err, state.ErrStateNotFound) {
				return err
			}

			return fmt.Errorf("get player state object failed (key=%s): %w", stateKey, err)
		}
		return nil
	}

	stateBytes, err := sc.manager.GetState(ctx, stateKey)
	if err != nil {
		if errors.Is(err, state.ErrStateNotFound) {
			return err
		}
		return fmt.Errorf("get player state failed (key=%s): %w", stateKey, err)
	}

	if err := json.Unmarshal(stateBytes, statePtr); err != nil {
		return fmt.Errorf("%w: failed to unmarshal player state (key=%s): %v", state.ErrDeserializationFailed, stateKey, err)
	}

	return nil
}

func (sc *StateClient) SetPlayerState(ctx context.Context, playerID uint64, stateObj interface{}, expirationSeconds int64) error {
	stateKey := fmt.Sprintf("player:%d", playerID)

	if setter, ok := sc.manager.(interface {
		SetStateObject(context.Context, string, interface{}, int64) error
	}); ok {
		err := setter.SetStateObject(ctx, stateKey, stateObj, expirationSeconds)
		if err != nil {
			return fmt.Errorf("set player state object failed (key=%s): %w", stateKey, err)
		}
		return nil
	}

	stateBytes, err := json.Marshal(stateObj)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal player state (key=%s): %v", state.ErrSerializationFailed, stateKey, err)
	}

	err = sc.manager.SetState(ctx, stateKey, stateBytes, expirationSeconds)
	if err != nil {
		return fmt.Errorf("set player state failed (key=%s): %w", stateKey, err)
	}

	return nil
}

func (sc *StateClient) DeletePlayerState(ctx context.Context, playerID uint64) error {
	stateKey := fmt.Sprintf("player:%d", playerID)
	err := sc.manager.DeleteState(ctx, stateKey)
	if err != nil {
		return fmt.Errorf("delete player state failed (key=%s): %w", stateKey, err)
	}
	return nil
}

func (sc *StateClient) GetRawManager() ziface.IStateManager {
	return sc.manager
}
