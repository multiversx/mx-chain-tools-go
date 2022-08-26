package mock

// CopyHandlerStub -
type CopyHandlerStub struct {
	CopyDirectoryCalled func(destination string, source string) error
}

// CopyDirectory -
func (stub *CopyHandlerStub) CopyDirectory(destination string, source string) error {
	if stub.CopyDirectoryCalled != nil {
		return stub.CopyDirectoryCalled(destination, source)
	}

	return nil
}

// IsInterfaceNil -
func (stub *CopyHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
