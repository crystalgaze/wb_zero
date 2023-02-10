package store

import (
	"log"
	"os"
	"strconv"
	"sync"
)

type Cache struct {
	buffer  map[int64]Order
	queue   []int64
	bufSize int
	pos     int
	DBInst  *Store
	name    string
	mutex   *sync.RWMutex
}

func NewCache(db *Store) *Cache {
	csh := Cache{}
	csh.Init(db)
	return &csh
}

func (c *Cache) Init(db *Store) {
	c.DBInst = db
	db.SetCahceInstance(c)
	c.name = "CACHE"
	c.mutex = &sync.RWMutex{}
	bufSize, err := strconv.Atoi(os.Getenv("CACHE_SIZE"))
	if err != nil {
		bufSize = 10
	}
	c.bufSize = bufSize
	c.buffer = make(map[int64]Order, c.bufSize)
	c.queue = make([]int64, c.bufSize)
	c.getCacheFromDatabase()
}

func (c *Cache) getCacheFromDatabase() {
	log.Printf("%v: поиск кэша в БД\n", c.name)
	buf, queue, pos, err := c.DBInst.GetCacheState(c.bufSize)
	if err != nil {
		log.Printf("%s: кэша нет %v\n", c.name, err)
		return
	}
	if pos == c.bufSize {
		pos = 0
	}
	c.mutex.Lock()
	c.buffer = buf
	c.queue = queue
	c.pos = pos
	c.mutex.Unlock()
	log.Printf("%s: кэш найден в БД     кэш %v", c.name, c.queue)
}

func (c *Cache) SetOrder(oid int64, o Order) {
	c.mutex.Lock()
	c.queue[c.pos] = oid
	c.pos++
	if c.pos == c.bufSize {
		c.pos = 0
	}
	c.buffer[oid] = o
	c.mutex.Unlock()
	c.DBInst.SendOrderIDToCache(oid)
	log.Printf("%s: order сохранен в кэш\n", c.name)
	log.Printf("%s: кэш  %v", c.name, c.queue)
}

func (c *Cache) GetOrderOutById(oid int64) (*OrderOut, error) {
	var out *OrderOut = &OrderOut{}
	var o Order
	var err error

	c.mutex.RLock()
	o, isExist := c.buffer[oid]
	c.mutex.RUnlock()

	if isExist {
		log.Printf("%s: id:%d взят из кэша\n", c.name, oid)
	} else {
		o, err = c.DBInst.GetOrderByID(oid)
		if err != nil {
			log.Printf("%s: ошибка orders: %v\n", c.name, err)
			return out, err
		}
		c.SetOrder(oid, o)
		log.Printf("%s: id:%d из orders сохранен в кэш \n", c.name, oid)
	}
	out.OrderUID = o.OrderUID
	out.TrackNumber = o.TrackNumber
	out.Entry = o.Entry
	out.Locale = o.Locale
	out.InternalSignature = o.InternalSignature
	out.CustomerID = o.CustomerID
	out.DeliveryService = o.DeliveryService
	out.Shardkey = o.Shardkey
	out.SmID = o.SmID
	out.DateCreated = o.DateCreated
	out.OofShard = o.OofShard
	return out, nil
}

func (c *Cache) Finish() {
	log.Printf("%s: завершение", c.name)
	c.DBInst.ClearCache()
	log.Printf("%s: завершенно", c.name)
}
