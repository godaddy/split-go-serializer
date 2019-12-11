package api

import "fmt"

// SplitioAPIBinding contains splitioAPIKey
type SplitioAPIBinding struct {
	splitioAPIKey string
}

// NewSplitioAPIBinding returns a new SplitioAPIBinding
func NewSplitioAPIBinding(splitioAPIKey string) *SplitioAPIBinding {
	return &SplitioAPIBinding{splitioAPIKey}
}

// HTTPGet will make a GET request to the Split.io SDK API
func (binding *SplitioAPIBinding) HTTPGet() error {
	return fmt.Errorf("not implemented")
}

// GetSegmentChanges will get segment data
func (binding *SplitioAPIBinding) GetSegmentChanges() error {
	return fmt.Errorf("not implemented")
}

// GetSplitChanges will get split data
func (binding *SplitioAPIBinding) GetSplitChanges() error {
	return fmt.Errorf("not implemented")
}
