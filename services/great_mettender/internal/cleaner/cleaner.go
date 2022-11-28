package cleaner

import (
	"context"
	"fmt"
	"time"

	"great_mettender/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func New(db *gorm.DB, ttl time.Duration) *Cleaner {
	return &Cleaner{
		ttl: ttl,
		db:  db,
	}
}

type Cleaner struct {
	ttl time.Duration
	db  *gorm.DB
}

func (c *Cleaner) Start(ctx context.Context) {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := c.collectOld(); err != nil {
				logrus.Errorf("Error collecting old tenders: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Cleaner) collectOld() error {
	res := c.db.
		Where("created_at < ?", time.Now().Add(-c.ttl)).
		Select(clause.Associations).
		Delete(&models.Tender{})

	if err := res.Error; err != nil {
		return fmt.Errorf("running query: %w", err)
	}

	logrus.Infof("Removed %d tenders", res.RowsAffected)

	return nil
}
