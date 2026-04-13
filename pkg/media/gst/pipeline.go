package gst

import (
	"fmt"

	"github.com/go-gst/go-gst/gst"
)

func newCapsFilter(caps string) (*gst.Element, error) {
	element, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, fmt.Errorf("create capsfilter: %w", err)
	}
	element.Set("caps", gst.NewCapsFromString(caps))
	return element, nil
}

func linkElements(elements ...*gst.Element) error {
	for i := 0; i < len(elements)-1; i++ {
		if err := elements[i].Link(elements[i+1]); err != nil {
			return fmt.Errorf("link %s -> %s: %w", elements[i].GetName(), elements[i+1].GetName(), err)
		}
	}
	return nil
}
