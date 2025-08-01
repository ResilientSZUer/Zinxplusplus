package scripting

import (
	"errors"
	"fmt"
	"sync"

	"zinxplusplus/ziface"

	lua "github.com/yuin/gopher-lua"
)

type LuaEngine struct {
	L        *lua.LState
	lock     sync.RWMutex
	isClosed bool
}

func NewLuaEngine() ziface.IScriptEngine {
	engine := &LuaEngine{
		L:        lua.NewState(lua.Options{}),
		isClosed: false,
	}

	engine.L.OpenLibs()

	fmt.Println("[Scripting] New LuaEngine created and initialized.")
	return engine
}

func (le *LuaEngine) Init() error {
	le.lock.Lock()
	defer le.lock.Unlock()
	if le.isClosed {
		return errors.New("lua engine is closed")
	}

	return nil
}

func (le *LuaEngine) LoadScripts(filePath string) error {
	le.lock.Lock()
	defer le.lock.Unlock()
	if le.isClosed {
		return errors.New("lua engine is closed")
	}

	fmt.Printf("[Scripting] Loading script file: %s\n", filePath)
	err := le.L.DoFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load script file '%s': %w", filePath, err)
	}
	fmt.Printf("[Scripting] Script file loaded successfully: %s\n", filePath)
	return nil
}

func (le *LuaEngine) CallFunc(funcName string, args ...interface{}) ([]interface{}, error) {
	le.lock.Lock()
	defer le.lock.Unlock()
	if le.isClosed {
		return nil, errors.New("lua engine is closed")
	}

	luaFunc := le.L.GetGlobal(funcName)
	if luaFunc.Type() != lua.LTFunction {
		return nil, fmt.Errorf("lua function '%s' not found or not a function", funcName)
	}

	luaArgs := make([]lua.LValue, len(args))
	for i, arg := range args {
		luaArgs[i] = GoValueToLValue(le.L, arg)
	}

	le.L.Push(luaFunc)
	for _, arg := range luaArgs {
		le.L.Push(arg)
	}

	err := le.L.PCall(len(luaArgs), lua.MultRet, nil)
	if err != nil {
		return nil, fmt.Errorf("error calling lua function '%s': %w", funcName, err)
	}

	numReturns := le.L.GetTop()
	results := make([]interface{}, numReturns)
	for i := 1; i <= numReturns; i++ {
		ret := le.L.Get(i)
		results[i-1] = LValueToGoValue(ret)
	}
	le.L.Pop(numReturns)

	return results, nil
}

func (le *LuaEngine) RegisterGoFunc(name string, goFunc interface{}) error {
	le.lock.Lock()
	defer le.lock.Unlock()
	if le.isClosed {
		return errors.New("lua engine is closed")
	}

	lgFunc, ok := goFunc.(func(*lua.LState) int)
	if !ok {

		return fmt.Errorf("failed to register '%s': function signature not directly compatible with lua.LGFunction (func(*lua.LState) int)", name)
	}

	le.L.SetGlobal(name, le.L.NewFunction(lgFunc))
	fmt.Printf("[Scripting] Registered Go function '%s' to Lua.\n", name)
	return nil
}

func (le *LuaEngine) Close() {
	le.lock.Lock()
	defer le.lock.Unlock()
	if !le.isClosed && le.L != nil {
		le.L.Close()
		le.isClosed = true
		fmt.Println("[Scripting] LuaEngine closed.")
	}
}

func GoValueToLValue(L *lua.LState, val interface{}) lua.LValue {
	if val == nil {
		return lua.LNil
	}
	switch v := val.(type) {
	case bool:
		return lua.LBool(v)
	case int:
		return lua.LNumber(v)
	case int32:
		return lua.LNumber(v)
	case int64:

		return lua.LNumber(v)
	case float32:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []byte:
		return lua.LString(string(v))
	case map[string]interface{}:
		return MapToLTable(L, v)
	case []interface{}:
		return SliceToLTable(L, v)

	default:
		panic(fmt.Sprintf("unsupported Go type to LValue: %T", val))
	}
}

func LValueToGoValue(lv lua.LValue) interface{} {
	switch lv.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(lv)
	case lua.LTNumber:
		return float64(lua.LVAsNumber(lv))
	case lua.LTString:
		return lua.LVAsString(lv)
	case lua.LTTable:
		return LTableToMap(lv.(*lua.LTable))

	default:

		return lv
	}
}

func MapToLTable(L *lua.LState, data map[string]interface{}) *lua.LTable {
	lt := L.NewTable()
	for k, v := range data {
		lt.RawSetString(k, GoValueToLValue(L, v))
	}
	return lt
}

func SliceToLTable(L *lua.LState, data []interface{}) *lua.LTable {
	lt := L.NewTable()
	for i, v := range data {
		lt.RawSetInt(i+1, GoValueToLValue(L, v))
	}
	return lt
}

func LTableToMap(lt *lua.LTable) map[string]interface{} {
	data := make(map[string]interface{})
	lt.ForEach(func(key lua.LValue, value lua.LValue) {
		if key.Type() == lua.LTString {
			data[key.String()] = LValueToGoValue(value)
		}

	})
	return data
}
