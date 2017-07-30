package miner

type cGroupDeleter interface {
	Delete() error
}
