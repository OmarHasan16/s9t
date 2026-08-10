package entity

import (
	"errors"
	"s9t.os/internal/modules/sales/crm/domain/valueobject"
	"s9t.os/internal/platform/types"
)

var (
	ErrDuplicateStagePosition = errors.New("pipeline cannot have duplicate stage positions")
	ErrStageNotFound          = errors.New("stage not found in this pipeline")
)

type Pipeline struct {
	ID        valueobject.PipelineID
	TenantID  types.TenantID
	Name      string
	IsDefault bool
	Stages    []PipelineStage
}

type PipelineStage struct {
	ID          valueobject.PipelineStageID
	PipelineID  valueobject.PipelineID
	Name        string
	Position    int
	Probability valueobject.Percentage
	IsWon       bool
	IsLost      bool
}

// AddStage ensures invariants like unique positions are maintained
func (p *Pipeline) AddStage(stage PipelineStage) error {
	for _, existing := range p.Stages {
		if existing.Position == stage.Position {
			return ErrDuplicateStagePosition
		}
	}
	p.Stages = append(p.Stages, stage)
	return nil
}

// HasStage verifies if a given stage belongs to this pipeline
func (p *Pipeline) HasStage(stageID valueobject.PipelineStageID) bool {
	for _, s := range p.Stages {
		if s.ID == stageID {
			return true
		}
	}
	return false
}
