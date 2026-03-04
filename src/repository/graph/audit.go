package graph

import (
	"context"
	"dpv/dpv/src/domain/entities"
	"encoding/json"
	"log"
	"os"
	"time"
)

type AuditAction string

const (
	ActionCreate AuditAction = "create"
	ActionUpdate AuditAction = "update"
	ActionDelete AuditAction = "delete"
)

type AuditEntry struct {
	Date   time.Time   `json:"date"`
	Author string      `json:"author"`
	Action AuditAction `json:"action"`
	Type   string      `json:"type"`
	Key    string      `json:"key"`
	Node   interface{} `json:"node,omitempty"`
}

type AuditLogger struct {
	FilePath string
}

func NewAuditLogger(filePath string) *AuditLogger {
	return &AuditLogger{FilePath: filePath}
}

func (l *AuditLogger) Log(ctx context.Context, action AuditAction, entityType string, key string, item interface{}) {
	if l == nil {
		return
	}

	author := "system"
	if user, ok := ctx.Value("user").(*entities.User); ok && user != nil {
		author = user.Key
	}

	// Scrub sensitive data conditionally from the memory
	loggableItem := item
	if filterable, ok := item.(entities.Filterable); ok {
		loggableItem = filterable.FilteredResponse(false)
	}

	entry := AuditEntry{
		Date:   time.Now(),
		Author: author,
		Action: action,
		Type:   entityType,
		Key:    key,
		Node:   loggableItem,
	}

	// For deletes, we might only have the key, so Node might be nil or partial
	if action == ActionDelete {
		entry.Node = nil
	}

	line, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal audit entry: %v", err)
		return
	}

	f, err := os.OpenFile(l.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open audit log: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(append(line, '\n')); err != nil {
		log.Printf("Failed to write audit log: %v", err)
	}
}
