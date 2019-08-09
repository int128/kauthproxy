package cmd

import (
	"testing"

	"github.com/go-test/deep"
)

func Test_parsePortPairNotation(t *testing.T) {
	tests := []struct {
		s    string
		want *portPair
	}{
		{"8443:https/443", &portPair{localPort: 8443, remoteScheme: "https", remotePort: 443}},
	}
	for _, test := range tests {
		t.Run(test.s, func(t *testing.T) {
			pair, err := parsePortPairNotation(test.s)
			if err != nil {
				t.Errorf("parsePortPairNotation error: %s", err)
			}
			if diff := deep.Equal(pair, test.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}
