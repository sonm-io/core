package network

import (
	"context"
	"fmt"
	"syscall"

	"github.com/sonm-io/core/insonmnia/worker/network/tc"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/vishvananda/netlink"
)

var _ sonm.QOSServer = &RemoteQOS{}

// NetworkAliasAction represents an action that is capable of creating an alias
// to the created bridge network.
//
// Must be appended just after creating Docker network.
type NetworkAliasAction struct {
	Network *Network
}

func (m *NetworkAliasAction) Execute(ctx context.Context) error {
	link, err := netlink.LinkByName(m.Network.Name)
	if err != nil {
		return err
	}

	return netlink.LinkSetAlias(link, m.Network.Alias)
}

func (m *NetworkAliasAction) Rollback() error {
	// The alias will be removed with the associated interface, so do nothing
	// here.
	return nil
}

type TBFShapingAction struct {
	Network         *Network
	tc              tc.TC
	rootQDiscHandle tc.QDisc
}

func (m *TBFShapingAction) Execute(ctx context.Context) error {
	link, err := netlink.LinkByName(m.Network.Name)
	if err != nil {
		return err
	}

	rootQDisc := &tc.TFBQDisc{
		QDiscAttrs: tc.QDiscAttrs{
			Link:   link,
			Handle: tc.NewHandle(0x8001, 0),
			Parent: tc.HandleRoot,
		},
		Rate:    m.Network.RateLimitIngress,
		Burst:   128000,
		Latency: 25000,
	}
	if err := m.tc.QDiscAdd(rootQDisc); err != nil {
		return fmt.Errorf("failed to add root queueing discipline: %s", err)
	}

	m.rootQDiscHandle = rootQDisc

	egressQDisc := &tc.Ingress{
		QDiscAttrs: tc.QDiscAttrs{
			Link:   link,
			Handle: tc.NewHandle(0xffff, 0),
			Parent: tc.HandleIngress,
		},
	}
	if err := m.tc.QDiscAdd(egressQDisc); err != nil {
		return fmt.Errorf("failed to add egress queueing discipline: %s", err)
	}

	filter := &tc.U32{
		FilterAttrs: tc.FilterAttrs{
			Link:     link,
			Parent:   egressQDisc.Handle,
			Priority: 1,
			Protocol: tc.ProtoIP,
		},
		FlowID: tc.NewHandle(0, 0),
		Actions: []tc.Action{
			&tc.PoliceAction{
				Rate:          m.Network.RateLimitEgress,
				Burst:         m.Network.RateLimitEgress,
				ConformExceed: tc.Drop,
			},
		},
	}
	if err := m.tc.FilterAdd(filter); err != nil {
		return fmt.Errorf("failed to add default filter for queueing discipline: %s", err)
	}

	return nil
}

func (m *TBFShapingAction) Rollback() error {
	if m.rootQDiscHandle != nil {
		return m.tc.QDiscDel(m.rootQDiscHandle)
	}

	return nil
}

type HTBShapingAction struct {
	Network         *Network
	tc              tc.TC
	rootQDiscHandle tc.QDisc
	ifbLink         netlink.Link
}

func (m *HTBShapingAction) Execute(ctx context.Context) error {
	if err := tc.IFBInit(); err != nil {
		return err
	}

	link, err := netlink.LinkByName(m.Network.Name)
	if err != nil {
		return err
	}

	rootQDisc := &tc.HTBQDisc{
		QDiscAttrs: tc.QDiscAttrs{
			Link:   link,
			Handle: tc.NewHandle(0x8001, 0),
			Parent: tc.HandleRoot,
		},
	}

	// Drop previous queueing discipline if it exists. This is the easiest way
	// to handle existing disciplines.
	m.tc.QDiscDel(rootQDisc)

	if err := m.tc.QDiscAdd(rootQDisc); err != nil {
		return fmt.Errorf("failed to add root queueing discipline: %s", err)
	}

	m.rootQDiscHandle = rootQDisc

	class := &tc.HTBClass{
		ClassAttrs: tc.ClassAttrs{
			Link:   link,
			Handle: rootQDisc.Handle.WithMinor(1),
			Parent: rootQDisc.Handle,
		},
		Rate: m.Network.RateLimitIngress,
		Ceil: m.Network.RateLimitIngress,
	}
	if err := m.tc.ClassAdd(class); err != nil {
		return fmt.Errorf("failed to add HTB class: %s", err)
	}

	leafQDisc := &tc.PfifoQDisc{
		QDiscAttrs: tc.QDiscAttrs{
			Link:   link,
			Handle: tc.NewHandle(0x8002, 0),
			Parent: class.Handle,
		},
		Limit: 1000,
	}
	if err := m.tc.QDiscAdd(leafQDisc); err != nil {
		return fmt.Errorf("failed to add leaf queueing discipline: %s", err)
	}

	filter := &tc.U32{
		FilterAttrs: tc.FilterAttrs{
			Link:     link,
			Parent:   rootQDisc.Handle,
			Priority: 1,
			Protocol: tc.ProtoAll,
		},
		FlowID: class.Handle,
		Selector: tc.U32Key{
			Val:  0x0,
			Mask: 0x0,
		},
		Actions: []tc.Action{},
	}
	if err := m.tc.FilterAdd(filter); err != nil {
		return fmt.Errorf("failed to add default filter for queueing discipline: %s", err)
	}

	// Configure traffic mirroring to apply egress shaping.
	ifbLink := m.newIFBLink()
	// NOTE: This won't be recreated after worker restart. No harm, I think, but who knows...
	if err := netlink.LinkAdd(ifbLink); err != nil && err != syscall.EEXIST {
		return fmt.Errorf("failed to add ifb device: %s", err)
	}
	if err := netlink.LinkSetAlias(ifbLink, fmt.Sprintf("%s%s", networkIfbPrefix, m.Network.Alias[len(networkPrefix):])); err != nil {
		return fmt.Errorf("failed to set alias to ifb device: %s", err)
	}
	if err := netlink.LinkSetUp(ifbLink); err != nil {
		return fmt.Errorf("failed to up ifb device: %s", err)
	}

	m.ifbLink = link

	egressQDisc := &tc.Ingress{
		QDiscAttrs: tc.QDiscAttrs{
			Link:   link,
			Handle: tc.NewHandle(0xffff, 0),
			Parent: tc.HandleIngress,
		},
	}
	if err := m.tc.QDiscAdd(egressQDisc); err != nil {
		return fmt.Errorf("failed to add egress queueing discipline: %s", err)
	}
	egressFilter := &tc.U32{
		FilterAttrs: tc.FilterAttrs{
			Link:     link,
			Parent:   egressQDisc.Handle,
			Protocol: tc.ProtoAll,
		},
		Selector: tc.U32Key{
			Val:  0x0,
			Mask: 0x0,
		},
		Actions: []tc.Action{
			&tc.MirredAction{
				Action: tc.EgressRedir,
				Dev:    ifbLink,
			},
		},
	}
	if err := m.tc.FilterAdd(egressFilter); err != nil {
		return fmt.Errorf("failed to add egress redirect filter: %s", err)
	}

	ifbQDisc := &tc.HTBQDisc{
		QDiscAttrs: tc.QDiscAttrs{
			Link:   ifbLink,
			Handle: tc.NewHandle(0x8001, 0),
			Parent: tc.HandleRoot,
		},
	}
	if err := m.tc.QDiscAdd(ifbQDisc); err != nil {
		return fmt.Errorf("failed to add ifb queueing discipline: %s", err)
	}

	ifbClass := &tc.HTBClass{
		ClassAttrs: tc.ClassAttrs{
			Link:   ifbLink,
			Handle: ifbQDisc.Handle.WithMinor(1),
			Parent: ifbQDisc.Handle,
		},
		Rate: m.Network.RateLimitEgress,
		Ceil: m.Network.RateLimitEgress,
	}
	if err := m.tc.ClassAdd(ifbClass); err != nil {
		return fmt.Errorf("failed to add ifb HTB class: %s", err)
	}

	ifbFilter := &tc.U32{
		FilterAttrs: tc.FilterAttrs{
			Link:     ifbLink,
			Parent:   ifbQDisc.Handle,
			Priority: 1,
			Protocol: tc.ProtoAll,
		},
		FlowID: ifbClass.Handle,
		Selector: tc.U32Key{
			Val:  0x0,
			Mask: 0x0,
		},
		Actions: []tc.Action{},
	}
	if err := m.tc.FilterAdd(ifbFilter); err != nil {
		return fmt.Errorf("failed to add ifb filter: %s", err)
	}

	return nil
}

func (m *HTBShapingAction) Rollback() error {
	errs := multierror.NewMultiError()

	if m.ifbLink == nil {
		m.ifbLink = m.newIFBLink()
	}

	if err := netlink.LinkDel(m.ifbLink); err != nil {
		errs = multierror.Append(errs, err)
	}
	if m.rootQDiscHandle != nil {
		if err := m.tc.QDiscDel(m.rootQDiscHandle); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

func (m *HTBShapingAction) newIFBLink() *netlink.Ifb {
	return NewIFBLink(m.Network.Name)
}

func NewIFBLink(name string) *netlink.Ifb {
	return &netlink.Ifb{
		LinkAttrs: netlink.LinkAttrs{
			Name:   fmt.Sprintf("%s%s", networkIfbPrefix, name[len(networkPrefix):]),
			TxQLen: 32,
		},
	}
}

func (m *localNetworkManager) Init() error {
	if err := tc.IFBFlush(); err != nil {
		return err
	}

	tcDefault, err := tc.NewDefaultTC()
	if err != nil {
		return err
	}

	m.tc = tcDefault

	return nil
}

func (m *localNetworkManager) NewActions(network *Network) []Action {
	return []Action{
		&DockerNetworkCreateAction{
			DockerClient: m.dockerClient,
			Network:      network,
		},
		&NetworkAliasAction{
			Network: network,
		},
		&HTBShapingAction{
			Network: network,
			tc:      m.tc,
		},
	}
}

type RemoteQOS struct {
	tc tc.TC
}

func NewRemoteQOS() (*RemoteQOS, error) {
	tcDefault, err := tc.NewDefaultTC()
	if err != nil {
		return nil, err
	}

	return &RemoteQOS{
		tc: tcDefault,
	}, nil
}

func (m *RemoteQOS) SetAlias(ctx context.Context, request *sonm.QOSSetAliasRequest) (*sonm.QOSSetAliasResponse, error) {
	action := NetworkAliasAction{
		Network: &Network{
			Name:  request.GetLinkName(),
			Alias: request.GetLinkAlias(),
		},
	}

	if err := action.Execute(ctx); err != nil {
		return nil, err
	}

	return &sonm.QOSSetAliasResponse{}, nil
}

func (m *RemoteQOS) AddHTBShaping(ctx context.Context, request *sonm.QOSAddHTBShapingRequest) (*sonm.QOSAddHTBShapingResponse, error) {
	action := HTBShapingAction{
		Network: &Network{
			Name:             request.GetLinkName(),
			Alias:            request.GetLinkAlias(),
			RateLimitEgress:  request.GetRateLimitEgress(),
			RateLimitIngress: request.GetRateLimitIngress(),
		},
		tc: m.tc,
	}

	if err := action.Execute(ctx); err != nil {
		return nil, err
	}

	return &sonm.QOSAddHTBShapingResponse{}, nil
}

func (m *RemoteQOS) RemoveHTBShaping(ctx context.Context, request *sonm.QOSRemoveHTBShapingRequest) (*sonm.QOSRemoveHTBShapingResponse, error) {
	linkName := request.GetLinkName()

	link, err := netlink.LinkByName(linkName)
	if err != nil {
		return nil, err
	}

	rootQDisc := &tc.HTBQDisc{
		QDiscAttrs: tc.QDiscAttrs{
			Link:   link,
			Handle: tc.NewHandle(0x8001, 0),
			Parent: tc.HandleRoot,
		},
	}

	action := HTBShapingAction{
		Network: &Network{
			Name: linkName,
		},
		tc:              m.tc,
		rootQDiscHandle: rootQDisc,
		ifbLink:         NewIFBLink(linkName),
	}
	action.ifbLink = action.newIFBLink()

	if err := action.Rollback(); err != nil {
		return nil, err
	}

	return &sonm.QOSRemoveHTBShapingResponse{}, nil
}

func (m *RemoteQOS) Flush(ctx context.Context, request *sonm.QOSFlushRequest) (*sonm.QOSFlushResponse, error) {
	if err := tc.IFBFlush(); err != nil {
		return nil, err
	}

	return &sonm.QOSFlushResponse{}, nil
}
