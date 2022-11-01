package models

import (
	"time"

	tenderspb "great_mettender/pkg/proto/tenders"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Tender struct {
	ID                 string    `gorm:"primaryKey;type:uuid"`
	Author             string    `gorm:"type:varchar(64)"`
	CreatedAt          time.Time `gorm:"index"`
	Winner             string    `gorm:"type:varchar(64)"`
	Name               string    `gorm:"type:varchar(64)"`
	Description        string    `gorm:"type:text"`
	Private            bool
	Finished           bool
	RequiredReputation float32
	ProgramInput       string `gorm:"type:text"`
}

func NewTenderFromProto(p *tenderspb.Tender) *Tender {
	return &Tender{
		ID:                 p.Id,
		Name:               p.Name,
		Description:        p.Description,
		RequiredReputation: p.RequiredReputation,
		Private:            p.Private,
		ProgramInput:       p.ProgramInput,
	}
}

func (t *Tender) ToProto(private, author bool) *tenderspb.Tender {
	return &tenderspb.Tender{
		Id:                 t.ID,
		Name:               t.Name,
		Description:        t.Description,
		RequiredReputation: t.RequiredReputation,
		Author:             t.formatPrivate(t.Author, private),
		Winner:             t.formatPrivate(t.Winner, private),
		Private:            t.Private,
		Finished:           t.Finished,
		ProgramInput:       t.formatPrivate(t.ProgramInput, author),
		CreatedAt:          timestamppb.New(t.CreatedAt),
	}
}

func (t *Tender) formatPrivate(s string, private bool) string {
	if !private && len(s) > 10 {
		return s[:10]
	}
	return s
}
