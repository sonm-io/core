package defergroup

// Just like "defer", but can be canceled. Useful in constructor functions when
// there are multiple resources need to be acquired, but that can fail to be
// allocated itself.
//
// Examples:
//  dg := DeferGroup{}
//	defer dg.Exec()
//
//	resource, err := Allocate()
//	if err != nil {
//		return err
//	}
//	dg.Defer(func() { resource.Close() })
//
//	// More allocations.
//
//	dg.CancelExec() // Everything is OK, no cleanup required.
//	return nil
type DeferGroup struct {
	fn       []func()
	canceled bool
}

func (m *DeferGroup) Defer(fn func()) {
	m.fn = append(m.fn, fn)
}

func (m *DeferGroup) Exec() {
	if m.canceled {
		return
	}

	for id := range m.fn {
		m.fn[len(m.fn)-1-id]()
	}
}

func (m *DeferGroup) CancelExec() {
	m.canceled = true
}
