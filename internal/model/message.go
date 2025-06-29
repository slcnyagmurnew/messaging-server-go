package model

type Message struct {
	Id      int
	Content string
	Phone   string
	Status  bool
}

type SentMessage struct {
	Content string `json:"content"`
	Phone   string `json:"to"`
}

type RedisMessage struct {
	MessageID   string `json:"messageId"`
	SendingTime string `json:"sendingTime"`
}
