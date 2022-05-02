package db

import (
	"errors"
	"fmt"

	"github.com/r04922101/portto/eth"
	"gorm.io/gorm"
)

// GetLatestNumFromDB gets largest block num in DB
func GetLatestNumFromDB(gdb *gorm.DB) (uint64, error) {
	var block *eth.Block
	if err := gdb.Last(&block).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to get latest block from DB: %v", err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return block.Num, nil
}
