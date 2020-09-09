package zproxy

import (
	"github.com/juju/fslock"
)

var lf = "/tmp/proxy_cfg.lock" // lock file

func Lock() *fslock.Lock {
	lock := fslock.New(lf)
	return lock
}

func Unlock(lock *fslock.Lock) {
	lock.Unlock()
}
