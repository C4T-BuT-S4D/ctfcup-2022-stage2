package models

import (
	"time"

	tenderspb "great_mettender/pkg/proto/tenders"
)

type Bid struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	TenderID    string `gorm:"type:uuid;index"`
	Description string `gorm:"type:text"`
	Program     string `gorm:"type:text"`
	Author      string `gorm:"type:varchar(64);index"`
	Won         bool
	Success     bool
	Price       float64
	CreatedAt   time.Time

	Tender *Tender `gorm:"foreignKey:TenderID;references:ID;constraint:OnDelete:CASCADE;"`
}

func NewBidFromProto(p *tenderspb.Bid) *Bid {
	return &Bid{
		ID:          p.Id,
		TenderID:    p.TenderId,
		Price:       p.Price,
		Description: p.Description,
		Program:     p.Program,
	}
}

func (b *Bid) ToProto() *tenderspb.Bid {
	return &tenderspb.Bid{
		Id:          b.ID,
		TenderId:    b.TenderID,
		Price:       b.Price,
		Description: b.Description,
		Program:     b.Program,
	}
}
