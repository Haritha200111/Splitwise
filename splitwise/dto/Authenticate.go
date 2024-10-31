package dto

import (
	"encoding/json"
	"time"
)

type RPCMessage struct {
	ServiceName     string              `json:"serviceName,omitempty"`
	MethodName      string              `json:"methodName,omitempty"`
	RequestName     string              `json:"requestName,omitempty"`
	MetaData        map[string][]string `json:"metaData,omitempty"`
	Error           string              `json:"error,omitempty"`
	ResponsePayload string              `json:"responsePayload,omitempty"`
	Status          string              `json:"status,omitempty"`
	RequestBody     []byte              `json:"requestBody,omitempty"`
	RequestTime     time.Time           `json:"requestTime,omitempty"`
	IpAddress       string              `json:"ipAddress,omitempty"`
}

type Params struct {
	Api     *Api            `json:"api,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Api struct {
	Credential string `json:"credential,omitempty"`
	Signature  string `json:"signature,omitempty"`
}

type AuthenticateRequest struct {
	ServiceName      string              `json:"serviceName,omitempty"`
	MethodName       string              `json:"methodName,omitempty"`
	RequestName      string              `json:"requestName,omitempty"`
	Application      string              `json:"application,omitempty"`
	RequestDigest    []byte              `json:"requestDigest,omitempty"`
	Authorization    string              `json:"authorization,omitempty"`
	RequestSignature string              `json:"requestSignature,omitempty"`
	UserAgent        string              `json:"userAgent,omitempty"`
	MetaData         map[string][]string `json:"metaData,omitempty"`
}
