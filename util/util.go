package util

import dto "github.com/prometheus/client_model/go"

// CreateLabelPair creates a label given a name and a value
func CreateLabelPair(name string, value string) *dto.LabelPair {
	return &dto.LabelPair{
		Name:  &name,
		Value: &value,
	}
}
