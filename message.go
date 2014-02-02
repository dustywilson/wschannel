package wschannel

import (
	"time"
)

type stateMessage struct {
	ServerMessage int           `json:"serverMessage"`
	Success       bool          `json:"success,omitempty"`
	Error         string        `json:"error,omitempty"`
	Disconnect    bool          `json:"disconnect,omitempty"`
	RetryDelay    time.Duration `json:"retryDelay,omitempty"`
	Setting       string        `json:"setting,omitempty"`
	Value         interface{}   `json:"value,omitempty"`
}

var MessageServiceOffline = stateMessage{
	Disconnect: true,
	Error:      "Service Offline",
}

var MessageServiceTemporarilyOffline = stateMessage{
	Disconnect: true,
	Error:      "Service Temporarily Offline",
	RetryDelay: time.Second * 20,
}

var MessageServiceRestarting = stateMessage{
	Disconnect: true,
	Error:      "Service Restarting",
	RetryDelay: time.Second * 5,
}

var MessageSessionEnded = stateMessage{
	Disconnect: true,
	Error:      "Session Ended",
	RetryDelay: time.Second * -1,
}

func NewSettingMessage(setting string, value interface{}) stateMessage {
	return stateMessage{
		Setting: setting,
		Value:   value,
	}
}
