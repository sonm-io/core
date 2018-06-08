package tc

import (
	"fmt"
	"os/exec"
)

const (
	TCActionAdd = "add"
	TCActionDel = "del"
)

// CMD represents "tc" commands and sub-commands that can be described using "tc"
type CMD interface {
	// Cmd creates and returns "tc" sub-command using this qdisc.
	Cmd() []string
}

type TC interface {
	QDiscAdd(qdisc QDisc) error
	QDiscDel(qdisc QDisc) error
	ClassAdd(class Class) error
	ClassDel(class Class) error
	FilterAdd(filter Filter) error
	FilterDel(filter Filter) error
	Close() error
}

type CmdTC struct {
	tcBin string
}

func NewCmdTC() (TC, error) {
	tcBin, err := exec.LookPath("tc")
	if err != nil {
		return nil, fmt.Errorf("failed to find `tc`: %s", err)
	}

	m := &CmdTC{
		tcBin: tcBin,
	}

	return m, nil
}

func (m *CmdTC) QDiscAdd(qdiscDesc QDisc) error {
	cmd := exec.Command(m.tcBin, append([]string{"qdisc", TCActionAdd, "dev", qdiscDesc.Attrs().Link.Attrs().Name}, qdiscDesc.Cmd()...)...)
	return cmd.Run()
}

func (m *CmdTC) QDiscDel(qdiscDesc QDisc) error {
	cmd := exec.Command(m.tcBin, append([]string{"qdisc", TCActionDel, "dev", qdiscDesc.Attrs().Link.Attrs().Name}, qdiscDesc.Cmd()...)...)
	return cmd.Run()
}

func (m *CmdTC) ClassAdd(class Class) error {
	cmd := exec.Command(m.tcBin, append([]string{"class", TCActionAdd, "dev", class.Attrs().Link.Attrs().Name}, class.Cmd()...)...)
	return cmd.Run()
}

func (m *CmdTC) ClassDel(class Class) error {
	cmd := exec.Command(m.tcBin, append([]string{"class", TCActionDel, "dev", class.Attrs().Link.Attrs().Name}, class.Cmd()...)...)
	return cmd.Run()
}

func (m *CmdTC) FilterAdd(filter Filter) error {
	cmd := exec.Command(m.tcBin, append([]string{"filter", TCActionAdd, "dev", filter.Attrs().Link.Attrs().Name}, filter.Cmd()...)...)
	return cmd.Run()
}

func (m *CmdTC) FilterDel(filter Filter) error {
	cmd := exec.Command(m.tcBin, append([]string{"filter", TCActionDel, "dev", filter.Attrs().Link.Attrs().Name}, filter.Cmd()...)...)
	return cmd.Run()
}

func (m *CmdTC) Close() error {
	return nil
}
