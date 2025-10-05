package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type App struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Nodes      []NodeModel      `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"nodes,omitempty"`
	Edges      []EdgeModel      `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"edges,omitempty"`
	GraphRuns  []GraphRunModel  `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"graph_runs,omitempty"`
}

type NodeModel struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	AppID       uuid.UUID `gorm:"type:char(36);not null;index" json:"app_id"`
	Type        string    `gorm:"type:varchar(50);not null;index" json:"type"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description,omitempty"`
	State       string    `gorm:"type:varchar(50);not null;default:'waiting';index" json:"state"`
	Properties  string    `gorm:"type:text;default:'{}'" json:"properties"` // JSON string (text for SQLite compatibility)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	App App `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"-"`
}

type EdgeModel struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	AppID       uuid.UUID `gorm:"type:char(36);not null;index" json:"app_id"`
	FromNodeID  string    `gorm:"not null;index" json:"from_node_id"`
	ToNodeID    string    `gorm:"not null;index" json:"to_node_id"`
	Type        string    `gorm:"type:varchar(50);not null;index" json:"type"`
	Description string    `json:"description,omitempty"`
	Properties  string    `gorm:"type:text;default:'{}'" json:"properties"` // JSON string (text for SQLite compatibility)
	CreatedAt   time.Time `json:"created_at"`

	App      App         `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"-"`
	FromNode NodeModel   `gorm:"foreignKey:FromNodeID;constraint:OnDelete:CASCADE" json:"-"`
	ToNode   NodeModel   `gorm:"foreignKey:ToNodeID;constraint:OnDelete:CASCADE" json:"-"`
}

type GraphRunModel struct {
	ID            uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	AppID         uuid.UUID  `gorm:"type:char(36);not null;index" json:"app_id"`
	Version       int        `gorm:"not null" json:"version"`
	Status        string     `gorm:"type:varchar(50);not null;default:'pending';index" json:"status"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	ExecutionPlan string     `gorm:"type:text" json:"execution_plan,omitempty"` // JSON string (text for SQLite compatibility)
	Metadata      string     `gorm:"type:text;default:'{}'" json:"metadata"`    // JSON string (text for SQLite compatibility)

	App App `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"-"`
}

func (App) TableName() string {
	return "graph_apps"
}

func (NodeModel) TableName() string {
	return "graph_nodes"
}

func (EdgeModel) TableName() string {
	return "graph_edges"
}

func (GraphRunModel) TableName() string {
	return "graph_runs"
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (gr *GraphRunModel) BeforeCreate(tx *gorm.DB) error {
	if gr.ID == uuid.Nil {
		gr.ID = uuid.New()
	}
	return nil
}