package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
)

type Store struct {
	pool *pgxpool.Pool
	csh  *Cache
	name string
}

func NewStore() *Store {
	db := Store{}
	db.Init()
	return &db
}

func (db *Store) SetCahceInstance(csh *Cache) {
	db.csh = csh
}

func (db *Store) GetCacheState(bufSize int) (map[int64]Order, []int64, int, error) {
	buffer := make(map[int64]Order, bufSize)
	queue := make([]int64, bufSize)
	var queueInd int
	query := fmt.Sprintf("SELECT order_id FROM cache WHERE app_key = '%s' ORDER BY id DESC LIMIT %d", os.Getenv("APP_KEY"), bufSize)
	rows, err := db.pool.Query(context.Background(), query)
	if err != nil {
		log.Printf("%v: не удалось получить Order_ID из БД %v\n", db.name, err)
	}
	defer rows.Close()
	var oid int64
	for rows.Next() {
		if err := rows.Scan(&oid); err != nil {
			log.Printf("%v: ошибка получения oid %v\n", db.name, err)
			return buffer, queue, queueInd, errors.New("ошибка получения oid")
		}
		queue[queueInd] = oid
		queueInd++
		o, err := db.GetOrderByID(oid)
		if err != nil {
			log.Printf("%v: не удалось получить order из БД %v\n", db.name, err)
			continue
		}
		buffer[oid] = o
	}
	if queueInd == 0 {
		return buffer, queue, queueInd, errors.New("кэш в БД пустой")
	}
	for i := 0; i < int(queueInd/2); i++ {
		queue[i], queue[queueInd-i-1] = queue[queueInd-i-1], queue[i]
	}
	return buffer, queue, queueInd, nil
}

func (db *Store) GetOrderByID(oid int64) (Order, error) {
	var o Order
	var payment_id_fk int64
	err := db.pool.QueryRow(context.Background(), `SELECT Order_UID, Entry, Internal_Signature, payment_fkey, Locale, Customer_ID, 
	Track_Number, Delivery_Service, Shardkey, Sm_ID FROM orders WHERE id = $1`, oid).Scan(&o.OrderUID, &o.Entry,
		&o.InternalSignature, &payment_id_fk, &o.Locale, &o.CustomerID, &o.TrackNumber, &o.DeliveryService, &o.Shardkey,
		&o.SmID)
	if err != nil {
		return o, errors.New("не удалось получить order из БД")
	}
	err = db.pool.QueryRow(context.Background(), `SELECT Transaction, Currency, Provider, Amount, Payment_Dt, Bank, Delivery_Cost,
	Goods_Total FROM payment WHERE id = $1`, payment_id_fk).Scan(&o.Payment.Transaction, &o.Payment.Currency, &o.Payment.Provider,
		&o.Payment.Amount, &o.Payment.PaymentDt, &o.Payment.Bank, &o.Payment.DeliveryCost, &o.Payment.GoodsTotal)
	if err != nil {
		log.Printf("%v: не удалось получить payment из БД %v\n", db.name, err)
		return o, errors.New("не удалось получить payment из БД")
	}
	rowsItems, err := db.pool.Query(context.Background(), "SELECT item_id_fk FROM order_items WHERE order_id_fk = $1", oid)
	if err != nil {
		return o, errors.New("не удалось получить items из БД")
	}
	defer rowsItems.Close()
	var itemID int64
	for rowsItems.Next() {
		var item Items
		if err := rowsItems.Scan(&itemID); err != nil {
			return o, errors.New("не удалось получить itemID из БД")
		}
		err = db.pool.QueryRow(context.Background(), `SELECT Chrt_ID, Price, Rid, Name, Sale, Size, Total_Price, Nm_ID, Brand 
		FROM items WHERE id = $1`, itemID).Scan(&item.ChrtID, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmID, &item.Brand)
		if err != nil {
			return o, errors.New("не удалось получить item из БД")
		}
		o.Items = append(o.Items, item)
	}
	return o, nil
}

func (db *Store) AddOrder(o Order) (int64, error) {
	var lastInsertId int64
	var itemsIds []int64 = []int64{}
	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())
	err = tx.QueryRow(context.Background(), `INSERT INTO delivery (name, phone, zip, city, address, region, email) values ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: не удалось сохранить delivery в БД %v\n", db.name, err)
		return -1, err
	}
	deliveryIdFk := lastInsertId
	err = tx.QueryRow(context.Background(), `INSERT INTO payment (transaction, currency, provider, amount, payment_dt, bank, delivery_cost,
		 goods_total, request_id, custom_fee) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`, o.Payment.Transaction, o.Payment.Currency, o.Payment.Provider,
		o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.RequestID, o.Payment.CustomFee).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: не удалось сохранить payment в БД %v\n", db.name, err)
		return -1, err
	}
	paymentIdFk := lastInsertId
	for _, item := range o.Items {
		err := tx.QueryRow(context.Background(), `INSERT INTO items (Chrt_ID, Price, Rid, Name, Sale, Size, Total_Price, Nm_ID, Brand, track_number, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`, item.ChrtID, item.Price, item.Rid, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.NmID, item.Brand, item.TrackNumber, item.Status).Scan(&lastInsertId)
		if err != nil {
			log.Printf("%v: не удалось сохранить items в БД %v\n", db.name, err)
			return -1, err
		}
		itemsIds = append(itemsIds, lastInsertId)
	}
	err = tx.QueryRow(context.Background(), `INSERT INTO orders (Order_UID, Entry, Internal_Signature, payment_fkey, Locale, 
		Customer_ID, Track_Number, Delivery_Service, Shardkey, Sm_ID, delivery_fkey, date_created, oof_shard) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`,
		o.OrderUID, o.Entry, o.InternalSignature, paymentIdFk, o.Locale, o.CustomerID, o.TrackNumber, o.DeliveryService,
		o.Shardkey, o.SmID, deliveryIdFk, o.DateCreated, o.OofShard).Scan(&lastInsertId)
	if err != nil {
		log.Printf("%v: не удалось сохранить orders в БД %v\n", db.name, err)
		return -1, err
	}
	orderIdFk := lastInsertId
	for _, itemId := range itemsIds {
		_, err := tx.Exec(context.Background(), `INSERT INTO order_items (order_id_fk, item_id_fk) values ($1, $2)`,
			orderIdFk, itemId)
		if err != nil {
			log.Printf("%v: не удалось сохранить order_items в БД %v\n", db.name, err)
			return -1, err
		}
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}
	log.Printf("%v: order сохранен в БД \n", db.name)
	db.csh.SetOrder(orderIdFk, o)
	return orderIdFk, nil
}
func (db *Store) SendOrderIDToCache(oid int64) {
	db.pool.QueryRow(context.Background(), `INSERT INTO cache (order_id, app_key) VALUES ($1, $2)`, oid, os.Getenv("APP_KEY"))
	log.Printf("%v: OrderID сохранен в кэш в БД\n", db.name)
}
func (db *Store) ClearCache() {
	_, err := db.pool.Exec(context.Background(), `DELETE FROM cache WHERE app_key = $1`, os.Getenv("APP_KEY"))
	if err != nil {
		log.Printf("%v: ошибка очистки кэша %s\n", db.name, err)
	}
	log.Printf("%v: кэш убран из БД\n", db.name)
}
