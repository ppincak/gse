package socket

type Store interface {

	Set(string, interface{})

	Get(string) interface{}

	Delete(string)

	DeleteAndGet(string) interface{}

	Has(string) bool

	Destroy()
}

type StoreFactory func() Store
