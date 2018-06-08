// +build linux,nl

package tc

// #cgo linux CFLAGS: -I /usr/include/libnl3/
// #cgo linux LDFLAGS: -lnl-3 -lnl-route-3
// #ifdef __linux__
//     #include <netlink/errno.h>
//     #include <netlink/netlink.h>
//     #include <netlink/socket.h>
//     #include <netlink/route/act/mirred.h>
//     #include <netlink/route/class.h>
//     #include <netlink/route/classifier.h>
//     #include <netlink/route/cls/u32.h>
//     #include <netlink/route/qdisc.h>
//     #include <netlink/route/qdisc/fifo.h>
//     #include <netlink/route/qdisc/htb.h>
//     #include <netlink/route/qdisc/tbf.h>
// #endif
import "C"

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

type NetlinkTC struct {
	sock *C.struct_nl_sock
}

// NewNetlinkTC constructs a new traffic control that works using netlink.
//
// It is the caller's responsibility to close returned TC.
func NewNetlinkTC() (TC, error) {
	sock := C.nl_socket_alloc()
	if ec := C.nl_connect(sock, C.NETLINK_ROUTE); ec != 0 {
		return nil, fmt.Errorf("failed to connect netlink socket: %s", C.GoString(C.nl_geterror(ec)))
	}

	m := &NetlinkTC{
		sock: sock,
	}

	return m, nil
}

func (m *NetlinkTC) QDiscAdd(qdiscDesc QDisc) error {
	return m.execQDisc(qdiscDesc, func(qdisc *C.struct_rtnl_qdisc) error {
		if ec := C.rtnl_qdisc_add(m.sock, qdisc, C.NLM_F_CREATE); ec != 0 {
			return fmt.Errorf("failed to add queueing discipline: %s", C.GoString(C.nl_geterror(ec)))
		}
		return nil
	})
}

func (m *NetlinkTC) QDiscDel(qdiscDesc QDisc) error {
	return m.execQDisc(qdiscDesc, func(qdisc *C.struct_rtnl_qdisc) error {
		if ec := C.rtnl_qdisc_delete(m.sock, qdisc); ec != 0 {
			return fmt.Errorf("failed to delete queueing discipline: %s", C.GoString(C.nl_geterror(ec)))
		}
		return nil
	})
}

func (m *NetlinkTC) execQDisc(qdiscDesc QDisc, fn func(qdisc *C.struct_rtnl_qdisc) error) error {
	qdisc := C.rtnl_qdisc_alloc()
	if qdisc == nil {
		return fmt.Errorf("failed to allocate qdisc")
	}
	defer C.rtnl_qdisc_put(qdisc)

	var nlCache *C.struct_nl_cache
	if ec := C.rtnl_link_alloc_cache(m.sock, unix.AF_UNSPEC, &nlCache); ec != 0 {
		return fmt.Errorf("failed to allocate link cache: %s", C.GoString(C.nl_geterror(ec)))
	}

	qdiscAttrs := qdiscDesc.Attrs()
	link := C.rtnl_link_get(nlCache, C.int(qdiscAttrs.Link.Attrs().Index))
	if link == nil {
		return fmt.Errorf("interface not found")
	}
	defer C.rtnl_link_put(link)

	C.rtnl_tc_set_link((*C.struct_rtnl_tc)(qdisc), link)
	C.rtnl_tc_set_parent((*C.struct_rtnl_tc)(qdisc), C.uint32_t(qdiscAttrs.Parent.UInt32()))
	C.rtnl_tc_set_handle((*C.struct_rtnl_tc)(qdisc), C.uint32_t(qdiscAttrs.Handle.UInt32()))

	qdiscName := C.CString(qdiscDesc.Type())
	defer C.free(unsafe.Pointer(qdiscName))

	if ec := C.rtnl_tc_set_kind((*C.struct_rtnl_tc)(qdisc), qdiscName); ec != 0 {
		return fmt.Errorf("failed to set qdisc kind: %s", C.GoString(C.nl_geterror(ec)))
	}

	switch v := qdiscDesc.(type) {
	case *PfifoQDisc:
		if ec := C.rtnl_qdisc_fifo_set_limit(qdisc, C.int(v.Limit)); ec != 0 {
			return fmt.Errorf("failed to set pfifo limit: %s", C.GoString(C.nl_geterror(ec)))
		}
	case *TFBQDisc:
		C.rtnl_qdisc_tbf_set_rate(qdisc, C.int(v.Rate), C.int(v.Burst), 0)
		if ec := C.rtnl_qdisc_tbf_set_limit_by_latency(qdisc, C.int(v.Latency)); ec != 0 {
			return fmt.Errorf("failed to set tbf latency: %s", C.GoString(C.nl_geterror(ec)))
		}
	case *HTBQDisc:
		// Nothing to configure here, except default class id. Ignore for now.
	case *Ingress:
		// Nothing to configure here.
	default:
		return fmt.Errorf("unknown qdisc type: %T", v)
	}

	return fn(qdisc)
}

func (m *NetlinkTC) ClassAdd(classDesc Class) error {
	class := C.rtnl_class_alloc()
	if class == nil {
		return fmt.Errorf("failed to allocate traffic control class")
	}
	defer C.rtnl_class_put(class)

	var nlCache *C.struct_nl_cache
	if ec := C.rtnl_link_alloc_cache(m.sock, unix.AF_UNSPEC, &nlCache); ec != 0 {
		return fmt.Errorf("failed to allocate link cache: %s", C.GoString(C.nl_geterror(ec)))
	}

	classAttrs := classDesc.Attrs()
	link := C.rtnl_link_get(nlCache, C.int(classAttrs.Link.Attrs().Index))
	if link == nil {
		return fmt.Errorf("interface not found")
	}
	defer C.rtnl_link_put(link)

	C.rtnl_tc_set_link((*C.struct_rtnl_tc)(class), link)
	C.rtnl_tc_set_parent((*C.struct_rtnl_tc)(class), C.uint32_t(classAttrs.Parent.UInt32()))
	C.rtnl_tc_set_handle((*C.struct_rtnl_tc)(class), C.uint32_t(classAttrs.Handle.UInt32()))

	className := C.CString(classDesc.Kind())
	defer C.free(unsafe.Pointer(className))

	if ec := C.rtnl_tc_set_kind((*C.struct_rtnl_tc)(class), className); ec != 0 {
		return fmt.Errorf("failed to set class kind: %s", C.GoString(C.nl_geterror(ec)))
	}

	switch v := classDesc.(type) {
	case *HTBClass:
		if v.Rate == 0 {
			v.Rate = 8
		}
		if v.Ceil == 0 {
			v.Ceil = 8
		}

		if ec := C.rtnl_htb_set_rate(class, C.uint32_t(v.Rate/8)); ec != 0 {
			return fmt.Errorf("failed to set class rate: %s", C.GoString(C.nl_geterror(ec)))
		}

		if ec := C.rtnl_htb_set_ceil(class, C.uint32_t(v.Ceil/8)); ec != 0 {
			return fmt.Errorf("failed to set class ceil: %s", C.GoString(C.nl_geterror(ec)))
		}

		if ec := C.rtnl_htb_set_rbuffer(class, C.uint32_t(1600)); ec != 0 {
			return fmt.Errorf("failed to set class burst: %s", C.GoString(C.nl_geterror(ec)))
		}

		if ec := C.rtnl_htb_set_cbuffer(class, C.uint32_t(1600)); ec != 0 {
			return fmt.Errorf("failed to set class cburst: %s", C.GoString(C.nl_geterror(ec)))
		}
	default:
		return fmt.Errorf("unknown class type: %T", v)
	}

	if ec := C.rtnl_class_add(m.sock, class, C.NLM_F_CREATE); ec != 0 {
		return fmt.Errorf("failed to add traffic control class: %s", C.GoString(C.nl_geterror(ec)))
	}

	return nil
}

func (m *NetlinkTC) ClassDel(class Class) error {
	return fmt.Errorf("unimplemented")
}

func (m *NetlinkTC) FilterAdd(filterDesc Filter) error {
	cls := C.rtnl_cls_alloc()
	if cls == nil {
		return fmt.Errorf("failed to allocate traffic control classifier")
	}
	defer C.rtnl_cls_put(cls)

	var nlCache *C.struct_nl_cache
	if ec := C.rtnl_link_alloc_cache(m.sock, unix.AF_UNSPEC, &nlCache); ec != 0 {
		return fmt.Errorf("failed to allocate link cache: %s", C.GoString(C.nl_geterror(ec)))
	}

	filterAttrs := filterDesc.Attrs()
	link := C.rtnl_link_get(nlCache, C.int(filterAttrs.Link.Attrs().Index))
	if link == nil {
		return fmt.Errorf("interface not found")
	}
	defer C.rtnl_link_put(link)

	C.rtnl_tc_set_link((*C.struct_rtnl_tc)(cls), link)

	filterKind := C.CString(filterDesc.Kind())
	defer C.free(unsafe.Pointer(filterKind))

	if ec := C.rtnl_tc_set_kind((*C.struct_rtnl_tc)(cls), filterKind); ec != 0 {
		return fmt.Errorf("failed to set filter kind: %s", C.GoString(C.nl_geterror(ec)))
	}

	if filterAttrs.Priority != 0 {
		C.rtnl_cls_set_prio(cls, C.uint16_t(filterAttrs.Priority))
	}
	C.rtnl_cls_set_protocol(cls, C.uint16_t(filterAttrs.Protocol))
	C.rtnl_tc_set_parent((*C.struct_rtnl_tc)(cls), C.uint32_t(filterAttrs.Parent.UInt32()))

	switch v := filterDesc.(type) {
	case *U32:
		if ec := C.rtnl_u32_add_key(cls, 0x0, 0x0, 0, 0); ec != 0 {
			return fmt.Errorf("failed to set filter key: %s", C.GoString(C.nl_geterror(ec)))
		}
		if v.FlowID != 0 {
			if ec := C.rtnl_u32_set_classid(cls, C.uint32_t(v.FlowID)); ec != 0 {
				return fmt.Errorf("failed to set filter flow id: %s", C.GoString(C.nl_geterror(ec)))
			}
		}
		if ec := C.rtnl_u32_set_cls_terminal(cls); ec != 0 {
			return fmt.Errorf("failed to set filter terminal: %s", C.GoString(C.nl_geterror(ec)))
		}

		for _, actionDesc := range v.Actions {
			fn := func(actionDesc Action) error {
				act := C.rtnl_act_alloc()
				if act == nil {
					return fmt.Errorf("failed to allocate traffic control action")
				}
				defer C.rtnl_act_put(act)

				actionKind := C.CString(actionDesc.Kind())
				defer C.free(unsafe.Pointer(actionKind))

				if ec := C.rtnl_tc_set_kind((*C.struct_rtnl_tc)(act), actionKind); ec != 0 {
					return fmt.Errorf("failed to set action kind: %s", C.GoString(C.nl_geterror(ec)))
				}

				switch a := actionDesc.(type) {
				case *MirredAction:
					if ec := C.rtnl_mirred_set_action(act, C.int(a.Action)); ec != 0 {
						return fmt.Errorf("failed to set action: %s", C.GoString(C.nl_geterror(ec)))
					}
					if ec := C.rtnl_mirred_set_ifindex(act, C.uint32_t(a.Dev.Attrs().Index)); ec != 0 {
						return fmt.Errorf("failed to set action if index: %s", C.GoString(C.nl_geterror(ec)))
					}
					if ec := C.rtnl_mirred_set_policy(act, 4); ec != 0 {
						return fmt.Errorf("failed to set action policy: %s", C.GoString(C.nl_geterror(ec)))
					}
				default:
					return fmt.Errorf("unknown action type: %T", a)
				}

				if ec := C.rtnl_u32_add_action(cls, act); ec != 0 {
					return fmt.Errorf("failed to add action: %s", C.GoString(C.nl_geterror(ec)))
				}

				return nil
			}

			if err := fn(actionDesc); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown filter type: %T", v)
	}

	// TODO: DO ACTIONS.

	if ec := C.rtnl_cls_add(m.sock, cls, C.NLM_F_CREATE); ec != 0 {
		return fmt.Errorf("failed to add traffic control filter: %s", C.GoString(C.nl_geterror(ec)))
	}

	return nil
}

func (m *NetlinkTC) FilterDel(filter Filter) error {
	return fmt.Errorf("unimplemented")
}

func (m *NetlinkTC) Close() error {
	C.nl_socket_free(m.sock)
	return nil
}

// NewDefaultTC constructs a new default traffic control.
func NewDefaultTC() (TC, error) {
	return NewNetlinkTC()
}
