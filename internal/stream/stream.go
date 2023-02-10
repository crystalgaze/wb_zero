package stream

import (
	"log"
	"os"
	"time"
	"wb_zero/internal/store"

	"github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

type Stream struct {
	conn  *stan.Conn
	sub   *Subscriber
	pub   *Publisher
	name  string
	isErr bool
}

func NewStream(db *store.Store) *Stream {
	sh := Stream{}
	sh.Init(db)
	return &sh
}

func (sh *Stream) Init(db *store.Store) {
	sh.name = "STREAM"
	err := sh.Connect()
	if err != nil {
		sh.isErr = true
		log.Printf("%s: ошибка %s", sh.name, err)
	} else {
		sh.sub = NewSubscriber(db, sh.conn)
		sh.sub.Subscribe()
		sh.pub = NewPublisher(sh.conn)
		sh.pub.Publish()
	}
}

func (sh *Stream) Connect() error {
	conn, err := stan.Connect(
		os.Getenv("NATS_CLUSTER_ID"),
		os.Getenv("NATS_CLIENT_ID"),
		stan.NatsURL(os.Getenv("NATS_HOSTS")),
		stan.NatsOptions(
			nats.ReconnectWait(time.Second*4),
			nats.Timeout(time.Second*4),
		),
		stan.Pings(7, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("%s: связь потеряна %v", sh.name, reason)
		}),
	)
	if err != nil {
		log.Printf("%s: не могу подключиться %v.\n", sh.name, err)
		return err
	}
	sh.conn = &conn
	log.Printf("%s: подключено", sh.name)
	return nil
}

func (sh *Stream) Finish() {
	if !sh.isErr {
		log.Printf("%s: Выключение", sh.name)
		sh.sub.Unsubscribe()
		(*sh.conn).Close()
		log.Printf("%s: Выключено", sh.name)
	}
}
