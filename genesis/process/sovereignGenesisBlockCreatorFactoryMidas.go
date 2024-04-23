package process

type sovereignGenesisBlockCreatorFactoryMidas struct {
}

// NewSovereignGenesisBlockCreatorFactory creates a sovereign genesis block creator factory
func NewSovereignGenesisBlockCreatorFactoryMidas() *sovereignGenesisBlockCreatorFactoryMidas {
	return &sovereignGenesisBlockCreatorFactoryMidas{}
}

// CreateGenesisBlockCreator creates a sovereign genesis block creator
func (gbf *sovereignGenesisBlockCreatorFactoryMidas) CreateGenesisBlockCreator(args ArgsGenesisBlockCreator) (GenesisBlockCreatorHandler, error) {
	gbc, err := NewGenesisBlockCreator(args)
	if err != nil {
		return nil, nil
	}

	return NewSovereignGenesisBlockCreatorMidas(gbc)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (gbf *sovereignGenesisBlockCreatorFactoryMidas) IsInterfaceNil() bool {
	return gbf == nil
}
