package cmd

import (
	"testing"

	"github.com/go-test/deep"
)

func Test_extractKubectlFlags(t *testing.T) {
	kubectlFlags, err := extractKubectlFlags([]string{"port-forward", "-n", "kube-system", "svc/foo"})
	if err != nil {
		t.Fatalf("extractKubectlFlags error: %+v", err)
	}
	if diff := deep.Equal(kubectlFlags, []string{"--namespace", "kube-system"}); diff != nil {
		t.Error(diff)
	}
}
