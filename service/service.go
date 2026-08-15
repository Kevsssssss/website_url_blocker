package service

import (
	"log"
	"time"

	"github.com/Kevsssssss/website_url_blocker/config"
	"github.com/kardianos/service"
)

// program implements the service.Interface for kardianos/service
type program struct {
	logger  service.Logger
	quit    chan struct{}
	lastMod time.Time
}

// NewService creates and returns a configured service.Service instance.
func NewService() (service.Service, error) {
	svcConfig := &service.Config{
		Name:        config.ServiceName,
		DisplayName: config.ServiceDisplayName,
		Description: config.ServiceDescription,
	}
	prg := &program{
		quit: make(chan struct{}),
	}
	return service.New(prg, svcConfig)
}

// Start is called when the service starts (non-blocking).
func (p *program) Start(s service.Service) error {
	p.logger, _ = s.Logger(nil)
	p.logInfo("URLBlocker service starting...")

	// Apply blocklist immediately on start
	if err := ApplyBlocklist(); err != nil {
		p.logError("Failed to apply blocklist on start: " + err.Error())
	} else {
		p.logInfo("Blocklist applied successfully.")
	}

	go p.run()
	return nil
}

// run is the background polling loop.
func (p *program) run() {
	ticker := time.NewTicker(config.PollInterval * time.Second)
	defer ticker.Stop()

	blocklistPath, err := config.BlocklistPath()
	if err != nil {
		p.logError("Cannot resolve blocklist path: " + err.Error())
		return
	}

	for {
		select {
		case <-ticker.C:
			p.checkAndReload(blocklistPath)
		case <-p.quit:
			return
		}
	}
}

// checkAndReload detects changes in blocklist.txt and hot-reloads if needed.
func (p *program) checkAndReload(blocklistPath string) {
	// No-op if we can't stat the file
	info, err := statFile(blocklistPath)
	if err != nil {
		return
	}
	if info.ModTime().After(p.lastMod) {
		p.logInfo("Blocklist changed, reloading...")
		if err := ApplyBlocklist(); err != nil {
			p.logError("Reload failed: " + err.Error())
		} else {
			p.lastMod = info.ModTime()
			p.logInfo("Blocklist reloaded successfully.")
		}
	}
}

// Stop is called when the service stops.
func (p *program) Stop(s service.Service) error {
	p.logInfo("URLBlocker service stopping, removing hosts entries...")
	close(p.quit)
	if err := RemoveBlocklist(); err != nil {
		p.logError("Failed to remove blocklist on stop: " + err.Error())
	}
	return nil
}

func (p *program) logInfo(msg string) {
	if p.logger != nil {
		_ = p.logger.Info(msg)
	} else {
		log.Println("[INFO]", msg)
	}
}

func (p *program) logError(msg string) {
	if p.logger != nil {
		_ = p.logger.Error(msg)
	} else {
		log.Println("[ERROR]", msg)
	}
}
