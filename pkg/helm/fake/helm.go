package fake

import (
	"bytes"
	"fmt"
)

// OCIPuller implements the ChartPuller interface
type OCIPuller struct {
	ExpectedName string
	Content      *bytes.Buffer
	Checksum     string
	Err          error
}

// PullOCIChart returns some fake content
func (f *OCIPuller) PullOCIChart(ociFullName string) (*bytes.Buffer, string, error) {
	if f.ExpectedName != "" && f.ExpectedName != ociFullName {
		return nil, "", fmt.Errorf("expecting %s got %s", f.ExpectedName, ociFullName)
	}
	return f.Content, f.Checksum, f.Err
}
