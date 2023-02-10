package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
	"wb_zero/internal/store"
)

type ordkey string

const orderKey ordkey = "order"

type Api struct {
	rtr                *chi.Mux
	csh                *store.Cache
	name               string
	srv                *http.Server
	httpServerExitDone *sync.WaitGroup
}

func NewApi(csh *store.Cache) *Api {
	api := Api{}
	api.Init(csh)
	return &api
}

func (a *Api) Init(csh *store.Cache) {
	a.csh = csh
	a.name = "API"
	a.rtr = chi.NewRouter()
	a.rtr.Get("/", a.WellcomeHandler)
	a.rtr.Route("/orders", func(r chi.Router) {
		r.Route("/{orderID}", func(r chi.Router) {
			r.Use(a.orderCtx)
			r.Get("/", a.GetOrder) // GET /orders/123
		})
	})
	a.httpServerExitDone = &sync.WaitGroup{}
	a.httpServerExitDone.Add(1)
	a.StartServer()
}

func (a *Api) Finish() {
	log.Printf("%v: Остановка api\n", a.name)
	if err := a.srv.Shutdown(context.Background()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
	a.httpServerExitDone.Wait()
	log.Printf("%v: Api остановлен\n", a.name)
}

func (a *Api) StartServer() {
	a.srv = &http.Server{
		Addr:    ":3333",
		Handler: a.rtr,
	}
	go func() {
		defer a.httpServerExitDone.Done()
		log.Printf("%v: адрес интерфейса http://localhost:3333\n", a.name)
		if err := a.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("error %v", err)
			return
		}
	}()
}

func (a *Api) orderCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderIDstr := chi.URLParam(r, "orderID")
		orderID, err := strconv.ParseInt(orderIDstr, 10, 64)
		if err != nil {
			log.Printf("%v: ошибка приведения %s в %v\n", a.name, orderIDstr, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		log.Printf("%v: запрос OrderOut из кеша/бд, OrderID: %v\n", a.name, orderIDstr)
		orderOut, err := a.csh.GetOrderOutById(orderID)
		if err != nil {
			log.Printf("%v: ошибка получения OrderOut из базы данных: %v\n", a.name, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), orderKey, orderOut)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Api) WellcomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/interface.html")
	if err != nil {
		log.Printf("%v: getOrder(): ошибка парсинга шаблона html: %s\n", a.name, err)
		http.Error(w, "Ошибка сервера ", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = t.ExecuteTemplate(w, "interface.html", nil)
	if err != nil {
		log.Printf("%v: WellcomeHandler(): ошибка выполнения шаблона html: %s\n", a.name, err)
		return
	}
}

func (a *Api) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderOut, ok := ctx.Value(orderKey).(*store.OrderOut)
	if !ok {
		log.Printf("%v: getOrder(): ошибка приведения интерфейса к типу *OrderOut\n", a.name)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}
	t, err := template.ParseFiles("web/interface.html")
	if err != nil {
		log.Printf("%v: getOrder(): ошибка парсинга шаблона html: %s\n", a.name, err)
		http.Error(w, "Ошибка сервера ", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	t.ExecuteTemplate(w, "interface.html", orderOut)
	if err != nil {
		log.Printf("%v: ошибка html %s\n", a.name, err)
		return
	}
}
