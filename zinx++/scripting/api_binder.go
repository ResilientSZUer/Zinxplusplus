package scripting

import (
	"encoding/json"
	"errors"
	"fmt"

	"zinxplusplus/ziface"

	lua "github.com/yuin/gopher-lua"
)

type ApiBinder struct {
	engine ziface.IScriptEngine
	server ziface.IServer
}

func NewApiBinder(engine ziface.IScriptEngine, server ziface.IServer) *ApiBinder {
	if engine == nil || server == nil {

		panic("engine and server must not be nil for ApiBinder")
	}
	return &ApiBinder{
		engine: engine,
		server: server,
	}
}

func (ab *ApiBinder) RegisterCoreAPI() error {
	fmt.Println("[ApiBinder] Registering Core Go APIs to Lua...")
	var errs []error

	if err := ab.engine.RegisterGoFunc("ZLogInfo", ab.luaLogInfo); err != nil {
		errs = append(errs, fmt.Errorf("register ZLogInfo: %w", err))
	}
	if err := ab.engine.RegisterGoFunc("ZLogError", ab.luaLogError); err != nil {
		errs = append(errs, fmt.Errorf("register ZLogError: %w", err))
	}

	if err := ab.engine.RegisterGoFunc("ZSendMsg", ab.luaSendMsg); err != nil {
		errs = append(errs, fmt.Errorf("register ZSendMsg: %w", err))
	}

	if err := ab.engine.RegisterGoFunc("ZGetConnProp", ab.luaGetConnProp); err != nil {
		errs = append(errs, fmt.Errorf("register ZGetConnProp: %w", err))
	}

	if err := ab.engine.RegisterGoFunc("ZSetConnProp", ab.luaSetConnProp); err != nil {
		errs = append(errs, fmt.Errorf("register ZSetConnProp: %w", err))
	}

	if len(errs) > 0 {

		errMsg := "Errors during API registration:\n"
		for _, e := range errs {
			errMsg += fmt.Sprintf("- %v\n", e)
		}
		return errors.New(errMsg)
	}

	fmt.Println("[ApiBinder] Core Go APIs registered successfully.")
	return nil
}

func (ab *ApiBinder) luaLogInfo(L *lua.LState) int {
	msg := L.ToString(1)
	fmt.Printf("[Lua Log Info] %s\n", msg)
	return 0
}

func (ab *ApiBinder) luaLogError(L *lua.LState) int {
	msg := L.ToString(1)
	fmt.Printf("[Lua Log Error] %s\n", msg)
	return 0
}

func (ab *ApiBinder) luaSendMsg(L *lua.LState) int {
	connID := L.ToNumber(1)
	msgID := L.ToNumber(2)
	msgTable := L.ToTable(3)

	if msgTable == nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("messageTable (3rd arg) cannot be nil"))
		return 2
	}

	goMap := LTableToMap(msgTable)
	msgBytes, err := json.Marshal(goMap)
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("failed to marshal messageTable: " + err.Error()))
		return 2
	}

	conn, err := ab.server.GetConnMgr().Get(uint64(connID))
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("connection not found: " + err.Error()))
		return 2
	}

	if err := conn.SendBuffMsg(uint32(msgID), msgBytes); err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("failed to send message: " + err.Error()))
		return 2
	}

	L.Push(lua.LBool(true))
	L.Push(lua.LNil)
	return 2
}

func (ab *ApiBinder) luaGetConnProp(L *lua.LState) int {
	connID := L.ToNumber(1)
	key := L.ToString(2)

	conn, err := ab.server.GetConnMgr().Get(uint64(connID))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("connection not found: " + err.Error()))
		return 2
	}

	value, err := conn.GetProperty(key)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("property not found or error: " + err.Error()))
		return 2
	}

	lValue := GoValueToLValue(L, value)
	L.Push(lValue)
	L.Push(lua.LNil)
	return 2
}

func (ab *ApiBinder) luaSetConnProp(L *lua.LState) int {
	connID := L.ToNumber(1)
	key := L.ToString(2)
	lValue := L.Get(3)

	conn, err := ab.server.GetConnMgr().Get(uint64(connID))
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("connection not found: " + err.Error()))
		return 2
	}

	goValue := LValueToGoValue(lValue)

	conn.SetProperty(key, goValue)

	L.Push(lua.LBool(true))
	L.Push(lua.LNil)
	return 2
}
