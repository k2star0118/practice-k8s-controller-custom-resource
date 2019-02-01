package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetKubernetesConfig(t *testing.T) {
	_, err := os.Stat(os.Getenv("HOME") + "/.kube/config")
	_, configErr := GetKubernetesConfig()
	if err == nil {
		assert.Nil(t, configErr)
	} else if os.IsNotExist(err) {
		assert.NotNil(t, configErr)
	}
}

func TestGetKubernetesClient(t *testing.T) {
	_, err := os.Stat(os.Getenv("HOME") + "/.kube/config")
	_, configErr := GetKubernetesClient()

	if err == nil {
		assert.Nil(t, configErr)
	} else if os.IsNotExist(err) {
		assert.NotNil(t, configErr)
	}
}

func TestGetMyKubernetesClient(t *testing.T) {
	_, err := os.Stat(os.Getenv("HOME") + "/.kube/config")
	_, configErr := GetMyKubernetesClient()

	if err == nil {
		assert.Nil(t, configErr)
	} else if os.IsNotExist(err) {
		assert.NotNil(t, configErr)
	}
}
