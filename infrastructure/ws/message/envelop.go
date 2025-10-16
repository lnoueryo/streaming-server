// infrastructure/ws/message/envelope.go
package message

type Envelope struct {
    Type string `json:"type"`
    Data any    `json:"data"`
}