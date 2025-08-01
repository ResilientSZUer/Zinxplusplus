package ziface

type IScriptEngine interface {
	Init() error

	LoadScripts(path string) error

	CallFunc(funcName string, args ...interface{}) ([]interface{}, error)

	RegisterGoFunc(name string, goFunc interface{}) error

	Close()
}
