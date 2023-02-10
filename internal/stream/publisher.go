package stream

import (
	"encoding/json"
	stan "github.com/nats-io/stan.go"
	"log"
	"os"
	"time"
	"wb_zero/internal/store"
)

type Publisher struct {
	sc   *stan.Conn
	name string
}

func NewPublisher(conn *stan.Conn) *Publisher {
	return &Publisher{
		name: "Publisher",
		sc:   conn,
	}
}

func (p *Publisher) Publish() {
	// публикация тестовых данных
	item1 := store.Items{ChrtID: 9934930, Price: 453, Rid: "ab4219087a764ae0btest", Name: "Shuba", Sale: 30, Size: "46", TotalPrice: 317, NmID: 2389212, Brand: "Gusi", TrackNumber: "WBILMTESTTRACK", Status: 202}
	item2 := store.Items{ChrtID: 9934930, Price: 4530, Rid: "hjkgdfhhjdf64ae0btest", Name: "Mascaras", Sale: 30, Size: "77", TotalPrice: 3170, NmID: 2389975, Brand: "Vivienne Sabo", TrackNumber: "WBILMTESTTRACK", Status: 202}
	item3 := store.Items{ChrtID: 9934930, Price: 7000, Rid: "gdfgbderera76ae0btest", Name: "Workboots", Sale: 10, Size: "53", TotalPrice: 6299, NmID: 2389723, Brand: "Bershka", TrackNumber: "WBILMTESTTRACK", Status: 202}
	delivery := store.Delivery{Name: "Serj Serjov", Phone: "+9005553535", Zip: "2639809", City: "Kiryat Mozkin", Address: "Ploshad Mira 15", Region: "Kraiot", Email: "test@gmail.com"}
	payment := store.Payment{Transaction: "b563feb7b2b84b6test", RequestID: "blank", Currency: "USD", Provider: "wbpay", CustomFee: 0, Amount: 1817, PaymentDt: 1637907727, Bank: "alpha", DeliveryCost: 1500, GoodsTotal: 317}
	order := store.Order{OrderUID: "Order 2", Entry: "2", InternalSignature: "blank", Delivery: delivery, Payment: payment, Items: []store.Items{item1, item2, item3},
		Locale: "en", CustomerID: "test", TrackNumber: "2", DeliveryService: "meest", Shardkey: "92", SmID: 99, DateCreated: time.Date(2021, 11, 26, 06, 22, 19, 0, time.UTC), OofShard: "1"}
	orderData, err := json.Marshal(order)
	if err != nil {
		log.Printf("%s: ошибка отправки тестовых данных %v\n", p.name, err)
	}
	ackHandler := func(ackedNuid string, err error) {
	}
	log.Printf("%s: публикация тестовых данных в канал %s \n", p.name, os.Getenv("NATS_SUBJECT"))
	nuid, err := (*p.sc).PublishAsync(os.Getenv("NATS_SUBJECT"), orderData, ackHandler)
	if err != nil {
		log.Printf("%s: ошибка публикации %s: %v\n", p.name, nuid, err.Error())
	} else {

	}
}
