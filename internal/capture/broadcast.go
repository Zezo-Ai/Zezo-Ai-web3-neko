package capture

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/demodesk/neko/server/pkg/gst"
	"gitlab.com/demodesk/neko/server/pkg/types"
)

type BroacastManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex

	pipeline    *gst.Pipeline
	pipelineStr string
	pipelineMu  sync.Mutex

	url     string
	started bool

	// metrics
	pipelinesCounter prometheus.Counter
}

func broadcastNew(pipelineStr string) *BroacastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "broadcast").
		Logger()

	return &BroacastManagerCtx{
		logger:      logger,
		pipelineStr: pipelineStr,
		url:         "",
		started:     false,

		// metrics
		pipelinesCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "pipelines_total",
			Namespace: "neko",
			Subsystem: "capture_broadcast",
			Help:      "Total number of created pipelines.",
		}),
	}
}

func (manager *BroacastManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Start(url string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.url = url
	manager.started = true
	return nil
}

func (manager *BroacastManagerCtx) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.started = false
	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Started() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.started
}

func (manager *BroacastManagerCtx) Url() string {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.url
}

func (manager *BroacastManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	var err error

	// replace {url} with valid URL
	pipelineStr := strings.Replace(manager.pipelineStr, "{url}", manager.url, 1)

	manager.logger.Info().
		Str("src", pipelineStr).
		Msgf("starting pipeline")

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.Play()
	manager.pipelinesCounter.Inc()

	return nil
}

func (manager *BroacastManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
