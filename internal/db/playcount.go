package db

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RecordPlay(userID uint, filePath string) error {
	now := time.Now()
	playCount := model.PlayCount{
		UserID:     userID,
		FilePath:   filePath,
		Count:      1,
		LastPlayed: now,
	}
	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "file_path"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"count":       gorm.Expr("count + 1"),
			"last_played": now,
		}),
	}).Create(&playCount).Error
	return errors.WithStack(err)
}

func GetPlayCounts(userID uint, filePaths []string) (map[string]*model.PlayCount, error) {
	var playCounts []model.PlayCount
	err := db.Where("user_id = ? AND file_path IN ?", userID, filePaths).Find(&playCounts).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res := make(map[string]*model.PlayCount)
	for i := range playCounts {
		res[playCounts[i].FilePath] = &playCounts[i]
	}
	return res, nil
}
