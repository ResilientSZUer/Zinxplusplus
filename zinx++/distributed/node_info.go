package distributed

import "fmt"

type NodeType string

const (
	NodeTypeGateway    NodeType = "gateway"
	NodeTypeGameServer NodeType = "gameserver"
	NodeTypeState      NodeType = "stateserver"
	NodeTypeManager    NodeType = "manager"
	NodeTypeUnknown    NodeType = "unknown"
)

type NodeInfo struct {
	NodeID   string   `json:"nodeId"`
	NodeType NodeType `json:"nodeType"`
	Host     string   `json:"host"`
	Port     int      `json:"port"`
}

func NewNodeInfo(nodeID string, nodeType NodeType, host string, port int) *NodeInfo {
	return &NodeInfo{
		NodeID:   nodeID,
		NodeType: nodeType,
		Host:     host,
		Port:     port,
	}
}

func (ni *NodeInfo) Addr() string {
	if ni == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", ni.Host, ni.Port)
}
