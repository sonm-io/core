// +build !linux !nl

package tc

func NewDefaultTC() (TC, error) {
	return NewCmdTC()
}
