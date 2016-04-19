package socket

type Store interface {

	Set(string, interface{})

	Get(string) interface{}

	Delete(string)

	Has(string) bool
}

type LocalStoreFactory func() *Store
