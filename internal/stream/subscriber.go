package stream

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
	"wb_zero/internal/store"

	stan "github.com/nats-io/stan.go"
)

type Subscriber struct {
	sub      stan.Subscription
	dbObject *store.Store
	sc       *stan.Conn
	name     string
}

func NewSubscriber(db *store.Store, conn *stan.Conn) *Subscriber {
	return &Subscriber{
		name:     "Subscriber",
		dbObject: db,
		sc:       conn,
	}
}

func (s *Subscriber) Subscribe() {
	var err error
	ackWait, err := strconv.Atoi(os.Getenv("NATS_ACK_WAIT_SECONDS"))
	if err != nil {
		log.Printf("%s: получено сообщение\n", s.name)
		return
	}
	s.sub, err = (*s.sc).Subscribe(
		os.Getenv("NATS_SUBJECT"),
		func(m *stan.Msg) {
			log.Printf("%s: получено сообщение\n", s.name)
			if s.messageHandler(m.Data) {
				err := m.Ack()
				if err != nil {
					log.Printf("%s ack ошибка %s", s.name, err)
				}
			}
		},
		stan.AckWait(time.Duration(ackWait)*time.Second),
		stan.DurableName(os.Getenv("NATS_DURABLE_NAME")),
		stan.SetManualAckMode(),
		stan.MaxInflight(10))
	if err != nil {
		log.Printf("%s: ошибка %v\n", s.name, err)
	}
	log.Printf("%s: создана подписка на канал %s\n", s.name, os.Getenv("NATS_SUBJECT"))
}

func (s *Subscriber) messageHandler(data []byte) bool {
	recievedOrder := store.Order{}
	err := json.Unmarshal(data, &recievedOrder)
	if err != nil {
		log.Printf("%s: ошибка в обработчике сообщений %v\n", s.name, err)
		return true
	}
	log.Printf("%s: приведение данных(json) к struct %v\n", s.name, recievedOrder)
	_, err = s.dbObject.AddOrder(recievedOrder)
	if err != nil {
		log.Printf("%s: не удалось добавить order %v\n", s.name, err)
		return false
	}
	return true
}

func (s *Subscriber) Unsubscribe() {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
}
