package controllers

import (
	"context"
	"fmt"

	"great_mettender/internal/models"

	"gorm.io/gorm"
)

func NewBids(ctx context.Context, db *gorm.DB) (*Bids, error) {
	b := &Bids{db: db}
	if err := b.initDB(ctx); err != nil {
		return nil, fmt.Errorf("initializing db: %w", err)
	}
	return b, nil
}

type Bids struct {
	db *gorm.DB
}

func (b *Bids) Add(ctx context.Context, bid *models.Bid) error {
	if err := b.db.WithContext(ctx).Create(bid).Error; err != nil {
		return fmt.Errorf("inserting bid: %w", err)
	}
	return nil
}

func (b *Bids) Get(ctx context.Context, id string) (*models.Bid, error) {
	var bid models.Bid
	if err := b.db.
		WithContext(ctx).
		Model(&models.Bid{}).
		Where("id = ?", id).
		First(&bid).
		Error; err != nil {
		return nil, fmt.Errorf("selecting bid: %w", err)
	}
	return &bid, nil
}

func (b *Bids) CalculateReputation(ctx context.Context, user string) (float32, error) {
	var res []struct {
		Result float32
	}
	if err := b.db.
		Model(&models.Bid{}).
		WithContext(ctx).
		Select("(count(*) filter (where success)) / greatest(count(*), 1) as result").
		Where("author = ?", user).
		Find(&res).
		Error; err != nil {
		return 0, fmt.Errorf("running query: %v", err)
	}
	if len(res) == 0 {
		return 0, nil
	}
	return res[0].Result, nil
}

func (b *Bids) ListByTender(ctx context.Context, tenderID string) ([]*models.Bid, error) {
	var bids []*models.Bid

	if err := b.db.
		Model(&models.Bid{}).
		WithContext(ctx).
		Where("tender_id = ?", tenderID).
		Order("created_at").
		Find(&bids).
		Error; err != nil {
		return nil, fmt.Errorf("running query: %v", err)
	}

	return bids, nil
}

func (b *Bids) SetWon(ctx context.Context, id string) error {
	if err := b.db.
		Model(&models.Bid{}).
		WithContext(ctx).
		Where("id = ?", id).
		UpdateColumn("won", true).
		Error; err != nil {
		return fmt.Errorf("running query: %v", err)
	}

	return nil
}

func (b *Bids) initDB(ctx context.Context) error {
	if err := b.db.WithContext(ctx).AutoMigrate(&models.Bid{}); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	return nil
}
