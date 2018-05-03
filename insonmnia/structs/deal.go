package structs

type DealID string

func (id DealID) String() string {
	return string(id)
}
