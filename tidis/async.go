//
// async.go
// Copyright (C) 2018 YanMing <yming0221@gmail.com>
//
// Distributed under terms of the MIT license.
//

package tidis

import "github.com/yongman/go/log"

type AsyncDelItem struct {
	keyType byte   // user key type
	ukey    []byte // user key
}

func (tidis *Tidis) AsyncDelAdd(keyType byte, ukey []byte) error {
	tidis.Lock.Lock()
	defer tidis.Lock.Unlock()

	key := string(keyType) + string(ukey)
	// key already added to chan queue
	if tidis.asyncDelSet.Contains(key) {
		return nil
	}
	tidis.asyncDelCh <- AsyncDelItem{keyType: keyType, ukey: ukey}
	tidis.asyncDelSet.Add(key)

	return nil
}

func (tidis *Tidis) AsyncDelDone(keyType byte, ukey []byte) error {
	tidis.Lock.Lock()
	defer tidis.Lock.Unlock()

	key := string(keyType) + string(ukey)
	if tidis.asyncDelSet.Contains(key) {
		tidis.asyncDelSet.Remove(key)
	}
	return nil
}

func (tidis *Tidis) RunAsync() {
	log.Infof("Async tasks started for async deletion")
	for {
		item := <-tidis.asyncDelCh
		tidis.AsyncDelDone(item.keyType, item.ukey)
		log.Debugf("Async recv key deletion %s", string(item.ukey))

		switch item.keyType {
		case TLISTMETA:
			deleted, err := tidis.Ldelete(item.ukey, false)
			if err != nil {
				log.Errorf("Async delete key %s error, %v", string(item.ukey), err.Error())
				continue
			}
			log.Debugf("async deletion key: %s result:%d", string(item.ukey), deleted)
		case THASHMETA:
		case TSETMETA:
		case TZSETMETA:
		}
	}
}