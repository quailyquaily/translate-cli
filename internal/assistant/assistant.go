package assistant

import (
	"sync"

	"github.com/lyricat/goutils/ai"
)

type (
	Assistant struct {
		cfg    Config
		aiInst *ai.Instant
		sync.Mutex
	}
	Config struct {
		Provider string
	}
)

func New(cfg Config, aiInst *ai.Instant) *Assistant {
	return &Assistant{
		cfg:    cfg,
		aiInst: aiInst,
	}
}
