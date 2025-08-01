package config

import (
	"zinxplusplus/state"
)

type Config struct {
	Server    ServerConfig    `json:"server"`
	Log       LogConfig       `json:"log"`
	State     StateConfig     `json:"state"`
	AOI       AOIConfig       `json:"aoi"`
	Scripting ScriptingConfig `json:"scripting"`
}

type ServerConfig struct {
	Name                   string `json:"name"`
	IPVersion              string `json:"ipVersion"`
	IP                     string `json:"ip"`
	Port                   int    `json:"port"`
	MaxConn                int    `json:"maxConn"`
	MaxPacketSize          uint32 `json:"maxPacketSize"`
	WorkerPoolSize         uint32 `json:"workerPoolSize"`
	MaxWorkerTaskLen       uint32 `json:"maxWorkerTaskLen"`
	ReadTimeoutMs          int    `json:"readTimeoutMs"`
	WriteTimeoutMs         int    `json:"writeTimeoutMs"`
	IdleTimeoutMs          int    `json:"idleTimeoutMs"`
	SendMsgTimeoutMs       int    `json:"sendMsgTimeoutMs"`
	SendTaskQueueTimeoutMs int    `json:"sendTaskQueueTimeoutMs"`
	MaxMsgChanLen          uint32 `json:"maxMsgChanLen"`
	MaxMsgBuffChanLen      uint32 `json:"maxMsgBuffChanLen"`
	NetpollNumLoops        int    `json:"netpollNumLoops"`
	NetpollLoadBalance     string `json:"netpollLoadBalance"`
}

type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputFile string `json:"outputFile"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	MaxAge     int    `json:"maxAge"`
	Compress   bool   `json:"compress"`
}

type StateConfig struct {
	Adapter string            `json:"adapter"`
	Redis   state.RedisConfig `json:"redis"`
}

type AOIConfig struct {
	MinX      float32 `json:"minX"`
	MaxX      float32 `json:"maxX"`
	MinZ      float32 `json:"minZ"`
	MaxZ      float32 `json:"maxZ"`
	CntsX     int     `json:"cntsX"`
	CntsZ     int     `json:"cntsZ"`
	Capacity  int     `json:"capacity"`
	MaxDepth  int     `json:"maxDepth"`
	ViewRange float32 `json:"viewRange"`
}

type ScriptingConfig struct {
	Enabled    bool   `json:"enabled"`
	EngineType string `json:"engineType"`
	ScriptPath string `json:"scriptPath"`
}
