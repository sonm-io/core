package hub

type HubsType struct {
	Id                  int
	Name                string
	TimeOfStart         int // TODO: cast to time.* Object
	AccountingPeriod    int
	Balance             float64
	MiddleSizeOfPayment float64
}