package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"zinxplusplus/state"
)

var GlobalConfig *Config

const DefaultConfigPath = "conf/zinxplusplus.json"

func init() {

	GlobalConfig = &Config{
		Server: ServerConfig{
			Name:                   "ZinxPlusServer-Default",
			IPVersion:              "tcp4",
			IP:                     "0.0.0.0",
			Port:                   8999,
			MaxConn:                1000,
			MaxPacketSize:          4096,
			WorkerPoolSize:         10,
			MaxWorkerTaskLen:       1024,
			ReadTimeoutMs:          30000,
			IdleTimeoutMs:          600000,
			SendMsgTimeoutMs:       3000,
			SendTaskQueueTimeoutMs: 100,
			MaxMsgChanLen:          1,
			MaxMsgBuffChanLen:      1024,
			NetpollNumLoops:        0,
			NetpollLoadBalance:     "round-robin",
		},
		Log: LogConfig{
			Level:      "debug",
			Format:     "text",
			OutputFile: "",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
		State: StateConfig{
			Adapter: "memory",
			Redis: state.RedisConfig{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
				PoolSize: 10,
			},
		},
		AOI: AOIConfig{
			MinX:      0,
			MaxX:      1000,
			MinZ:      0,
			MaxZ:      1000,
			Capacity:  4,
			MaxDepth:  8,
			ViewRange: 50,
			CntsX:     10,
			CntsZ:     10,
		},
		Scripting: ScriptingConfig{
			Enabled:    false,
			EngineType: "lua",
			ScriptPath: "./scripts",
		},
	}
}

func InitGlobalConfig(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = DefaultConfigPath
	}
	fmt.Printf("[Config] Initializing global config from: %s\n", configFilePath)

	return LoadConfig(configFilePath)
}

func LoadConfig(filePath string) error {
	fmt.Printf("[Config] Attempting to load config file: %s\n", filePath)

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for config file '%s': %w", filePath, err)
	}
	fmt.Printf("[Config] Absolute config path: %s\n", absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("[Config Warning] Config file '%s' not found. Using default configuration set during init.\n", absPath)

			if GlobalConfig == nil {

				panic("GlobalConfig is nil after init, cannot proceed without default config.")
			}
			return nil
		}
		return fmt.Errorf("error reading config file '%s': %w", absPath, err)
	}

	configHolder := *GlobalConfig

	err = json.Unmarshal(data, &configHolder)
	if err != nil {
		return fmt.Errorf("error unmarshalling config file '%s': %w", absPath, err)
	}

	GlobalConfig = &configHolder

	fmt.Printf("[Config] Config loaded successfully from '%s'.\n", absPath)
	return nil
}
