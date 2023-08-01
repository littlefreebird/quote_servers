package sample

import (
	"errors"
	"quote/common/log"
	"testing"
)

func TestNoConfig(t *testing.T) {
	log.Info("hello world")
	log.Infof("hello world, %s and %s", "ABC", "123")
	log.Warnf("warn info: %s", "wrong")
	log.Errorf("error info: %v", errors.New("crash crash"))
}
