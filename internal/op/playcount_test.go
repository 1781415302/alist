package op

import (
	"sync"
	"testing"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	dB, err := gorm.Open(sqlite.Open("file:memdb_playcount?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	conf.Conf = conf.DefaultConfig()
	db.Init(dB)
}

func TestPlayCount(t *testing.T) {
	userID := uint(42)
	filePath := "/test/video.mp4"

	// Clear memory map
	lastPlayedMap.Delete("42:/test/video.mp4")

	// Clear database record to ensure test repeatability
	db.GetDb().Where("user_id = ? AND file_path = ?", userID, filePath).Delete(&model.PlayCount{})

	// 1. Initial count query
	counts, err := GetPlayCounts(userID, []string{filePath})
	if err != nil {
		t.Fatalf("failed to query play counts: %+v", err)
	}
	if val, ok := counts[filePath]; ok {
		t.Errorf("expected no record, got count %d", val.Count)
	}

	// 2. Report play first time
	err = ReportPlay(userID, filePath)
	if err != nil {
		t.Fatalf("failed to report play count: %+v", err)
	}

	counts, err = GetPlayCounts(userID, []string{filePath})
	if err != nil {
		t.Fatalf("failed to query play counts: %+v", err)
	}
	if val, ok := counts[filePath]; !ok || val.Count != 1 {
		t.Errorf("expected count 1, got %+v", val)
	}

	// 3. Report play second time immediately (should be deduplicated)
	err = ReportPlay(userID, filePath)
	if err != nil {
		t.Fatalf("failed to report play count second time: %+v", err)
	}

	counts, err = GetPlayCounts(userID, []string{filePath})
	if err != nil {
		t.Fatalf("failed to query play counts second time: %+v", err)
	}
	if val, ok := counts[filePath]; !ok || val.Count != 1 {
		t.Errorf("expected count to remain 1 due to 5s deduplication, got %+v", val)
	}

	// 4. Simulate time elapsed by manually clearing/modifying map
	lastPlayedMap.Delete("42:/test/video.mp4")

	err = ReportPlay(userID, filePath)
	if err != nil {
		t.Fatalf("failed to report play count after 5s: %+v", err)
	}

	counts, err = GetPlayCounts(userID, []string{filePath})
	if err != nil {
		t.Fatalf("failed to query play counts after 5s: %+v", err)
	}
	if val, ok := counts[filePath]; !ok || val.Count != 2 {
		t.Errorf("expected count to be 2 after simulating 5s elapsed, got %+v", val)
	}
}

func TestPlayCountConcurrent(t *testing.T) {
	userID := uint(99)
	filePath := "/test/concurrent_video.mp4"

	// Clear DB and map
	lastPlayedMap.Delete("99:/test/concurrent_video.mp4")
	db.GetDb().Where("user_id = ? AND file_path = ?", userID, filePath).Delete(&model.PlayCount{})

	// Fire 50 concurrent goroutines to report play count
	var wg sync.WaitGroup
	concurrentCount := 50
	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ReportPlay(userID, filePath); err != nil {
				t.Errorf("ReportPlay failed concurrently: %+v", err)
			}
		}()
	}
	wg.Wait()

	// Query play count, should be exactly 1 because all were concurrent within 5 seconds
	counts, err := GetPlayCounts(userID, []string{filePath})
	if err != nil {
		t.Fatalf("failed to query play counts: %+v", err)
	}
	val, ok := counts[filePath]
	if !ok {
		t.Fatalf("expected play count record to exist")
	}
	if val.Count != 1 {
		t.Errorf("expected play count to be exactly 1 under 50 concurrent reports, got %d", val.Count)
	}
}
