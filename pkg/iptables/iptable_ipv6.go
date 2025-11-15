package iptables

// NewIPV6 returns a command wrapper configured for ip6tables binaries.
func NewIPV6(opt ...option) (*IptablesV4CMD, error) {
	options := append([]option{}, opt...)
	options = append(options, WithProtocol(ProtocolIPv6))
	return newIptablesCommand(ProtocolIPv6, options...)
}
