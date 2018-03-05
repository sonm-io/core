package network

type octet struct {
	occupied byte
	subnodes map[int]*octet
}
