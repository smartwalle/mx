package mx

type Message interface {
	Value() []byte

	Ack() error
}
