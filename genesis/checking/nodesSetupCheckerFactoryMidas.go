package checking

type nodesSetupCheckerFactoryMidas struct {
}

// NewNodesSetupCheckerFactory creates a nodes setup checker factory
func NewNodesSetupCheckerFactoryMidas() *nodesSetupCheckerFactoryMidas {
	return &nodesSetupCheckerFactoryMidas{}
}

// CreateNodesSetupChecker creates a new nodes setup checker
func (f *nodesSetupCheckerFactoryMidas) CreateNodesSetupChecker(args ArgsNodesSetupChecker) (NodesSetupChecker, error) {
	return NewNodesSetupCheckerMidas(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *nodesSetupCheckerFactoryMidas) IsInterfaceNil() bool {
	return f == nil
}
