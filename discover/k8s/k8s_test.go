package k8s_test

import (
	"context"
	"encoding/json"
	"github.com/graphikDB/graphik/discover/k8s"
	"testing"
)

func Test(t *testing.T) {
	provider, err := k8s.NewOutOfClusterProvider("default")
	if err != nil {
		t.Fatal(err.Error())
	}
	pods, err := provider.Pods(context.Background())
	if err != nil {
		t.Fatal(err.Error())
	}
	bits, err := json.MarshalIndent(pods, "", "    ")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(string(bits))
}
