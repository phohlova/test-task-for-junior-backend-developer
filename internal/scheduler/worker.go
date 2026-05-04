package scheduler

import (
	"context"
	"fmt"
	"time"

	"example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/logger"
)

type WorkerConfig struct {
	TickInterval time.Duration
	Lookahead
}