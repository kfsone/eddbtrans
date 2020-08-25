package eddbtrans

type Message struct {
	ObjectId uint64 // unique identifier for the entry
	Data     []byte // the data of the message
}
