package op

import (
	"strconv"
	"sync"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
)

var (
	lastPlayedMap sync.Map
)

func ReportPlay(userID uint, filePath string) error {
	key := strconv.FormatUint(uint64(userID), 10) + ":" + filePath
	for {
		now := time.Now()
		val, loaded := lastPlayedMap.Load(key)
		if !loaded {
			actual, stored := lastPlayedMap.LoadOrStore(key, now)
			if !stored {
				time.AfterFunc(5*time.Second, func() {
					lastPlayedMap.CompareAndDelete(key, now)
				})
				err := db.RecordPlay(userID, filePath)
				if err != nil {
					lastPlayedMap.CompareAndDelete(key, now)
				}
				return err
			}
			val = actual
		}

		lastTime, ok := val.(time.Time)
		if !ok {
			lastPlayedMap.Delete(key)
			continue
		}

		if now.Sub(lastTime) < 5*time.Second {
			return nil
		}

		if lastPlayedMap.CompareAndSwap(key, lastTime, now) {
			time.AfterFunc(5*time.Second, func() {
				lastPlayedMap.CompareAndDelete(key, now)
			})
			err := db.RecordPlay(userID, filePath)
			if err != nil {
				lastPlayedMap.CompareAndDelete(key, now)
			}
			return err
		}
	}
}

func GetPlayCounts(userID uint, filePaths []string) (map[string]*model.PlayCount, error) {
	return db.GetPlayCounts(userID, filePaths)
}
