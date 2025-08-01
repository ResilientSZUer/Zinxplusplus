package scripting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"zinxplusplus/ziface"
)

type ScriptManager struct {
	engine  ziface.IScriptEngine
	binder  *ApiBinder
	scripts map[string]bool
	lock    sync.RWMutex
	server  ziface.IServer
}

func NewScriptManager(server ziface.IServer) (*ScriptManager, error) {
	engine := NewLuaEngine()
	binder := NewApiBinder(engine, server)

	sm := &ScriptManager{
		engine:  engine,
		binder:  binder,
		scripts: make(map[string]bool),
		server:  server,
	}

	if err := engine.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize script engine: %w", err)
	}

	if err := sm.binder.RegisterCoreAPI(); err != nil {

		fmt.Printf("[ScriptManager] Warning: Failed to register some core APIs: %v\n", err)
	}

	fmt.Println("[ScriptManager] ScriptManager created.")
	return sm, nil
}

func (sm *ScriptManager) LoadScript(filePath string) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if err := sm.engine.LoadScripts(filePath); err != nil {
		return err
	}
	sm.scripts[filePath] = true
	return nil
}

func (sm *ScriptManager) LoadScriptDir(dirPath string) error {
	fmt.Printf("[ScriptManager] Loading scripts from directory: %s\n", dirPath)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read script directory '%s': %w", dirPath, err)
	}

	loadedCount := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".lua") {
			filePath := filepath.Join(dirPath, file.Name())
			if err := sm.LoadScript(filePath); err != nil {

				fmt.Printf("[ScriptManager] Error loading script '%s': %v\n", filePath, err)

				continue
			}
			loadedCount++
		}
	}
	fmt.Printf("[ScriptManager] Finished loading scripts from '%s'. Loaded %d files.\n", dirPath, loadedCount)
	return nil
}

func (sm *ScriptManager) Call(funcName string, args ...interface{}) ([]interface{}, error) {

	return sm.engine.CallFunc(funcName, args...)
}

func (sm *ScriptManager) GetEngine() ziface.IScriptEngine {
	return sm.engine
}

func (sm *ScriptManager) Close() {
	fmt.Println("[ScriptManager] Closing ScriptManager...")
	sm.engine.Close()
	fmt.Println("[ScriptManager] ScriptManager closed.")
}
