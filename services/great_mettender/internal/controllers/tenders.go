package controllers

import (
	"context"
	"fmt"

	"great_mettender/internal/models"

	"gorm.io/gorm"
)

func NewTenders(ctx context.Context, db *gorm.DB) (*Tenders, error) {
	t := &Tenders{db: db}
	if err := t.initDB(ctx); err != nil {
		return nil, fmt.Errorf("initializing db: %w", err)
	}
	return t, nil
}

type Tenders struct {
	db *gorm.DB
}

func (t *Tenders) Add(ctx context.Context, tender *models.Tender) error {
	if err := t.db.WithContext(ctx).Create(tender).Error; err != nil {
		return fmt.Errorf("inserting tender: %w", err)
	}
	return nil
}

func (t *Tenders) Get(ctx context.Context, id string) (*models.Tender, error) {
	var tender models.Tender
	if err := t.db.
		WithContext(ctx).
		Model(&models.Tender{}).
		Where("id = ?", id).
		First(&tender).
		Error; err != nil {
		return nil, fmt.Errorf("selecting tender: %w", err)
	}
	return &tender, nil
}

func (t *Tenders) Finish(ctx context.Context, id string) error {
	if err := t.db.
		WithContext(ctx).
		Model(&models.Tender{}).
		Where("id = ?", id).
		UpdateColumn("finished", true).
		Error; err != nil {
		return fmt.Errorf("running query: %w", err)
	}
	return nil
}

func (t *Tenders) initDB(ctx context.Context) error {
	if err := t.db.WithContext(ctx).AutoMigrate(&models.Tender{}); err != nil {
		return fmt.Errorf("migrating db: %w", err)
	}
	return nil
}
