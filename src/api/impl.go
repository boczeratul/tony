package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/r04922101/portto/eth"
	"github.com/r04922101/portto/indexer"
	"gorm.io/gorm"
)

const (
	defaultLimit = 20
)

type serviceImpl struct {
	db        *gorm.DB
	ethClient eth.Client
	indexer   indexer.Indexer
}

func (s *serviceImpl) getBlocks(c *gin.Context) {
	l := c.Query("limit")
	limit, err := strconv.Atoi(l)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad limit query parameter"))
		return
	}
	if limit <= 0 {
		limit = defaultLimit
	}

	// get blocks from DB
	var blocks []*eth.Block
	if err := s.db.Debug().Model(&eth.Block{}).Preload("Transactions").
		Order("num desc").Limit(limit).
		Find(&blocks).Error; err != nil {
		log.Printf("failed to find recent %d blocks from DB: %v", limit, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	for _, b := range blocks {
		toRepsonseBlock(b)
	}

	// index new blocks to DB in background
	go func(blocks []*eth.Block) {
		// get latest number from DB
		var start uint64
		if len(blocks) == 0 {
			curNum, err := s.ethClient.GetCurrentNumber(context.Background())
			if err != nil {
				return
			}
			start = curNum
		} else {
			start = blocks[0].Num + 1
		}

		// index new blocks into DB
		if _, err := s.indexer.IndexRecentBlocks(context.Background(), start); err != nil {
			log.Printf("failed to index recent blocks to DB: %v", err)
		}
	}(blocks)

	c.JSON(http.StatusOK, blocks)
}

func (s *serviceImpl) getBlockByHash(c *gin.Context) {
	h := c.Param("id")
	if h == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad id path parameter"))
		return
	}

	var (
		block *eth.Block
		err   error
	)
	// try db exists
	if err = s.db.Where("hash = ?", h).Find(&block).Error; err != nil {
		log.Printf("failed to find block with hash %s in DB: %v", h, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if block.Hash != h {
		// get from RPC and index it into DB
		ctx := c.Request.Context()
		block, err = s.ethClient.GetBlockByHash(ctx, h)
		if err != nil {
			log.Printf("failed to call RPC get block by hash %s: %v", h, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// write block into DB in background
		go func() {
			s.indexer.IndexBlock(block)
		}()
	}

	toRepsonseBlock(block)
	c.JSON(http.StatusOK, block)
}

func (s *serviceImpl) getTransactionByHash(c *gin.Context) {
	h := c.Param("txHash")
	if h == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad txHash path parameter"))
		return
	}

	var (
		tx  *eth.Transaction
		err error
	)
	// try db exists
	if err = s.db.Preload("Logs").Where("hash = ?", h).Find(&tx).Error; err != nil {
		log.Printf("failed to find transaction with hash %s in DB: %v", h, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if tx.Hash != h {
		// get from RPC
		ctx := c.Request.Context()
		tx, err = s.ethClient.GetTransactionByHash(ctx, h)
		if err != nil {
			log.Printf("failed to call RPC get transaction by hash %s: %v", h, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, tx)
}
