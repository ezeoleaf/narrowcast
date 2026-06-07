// narrowcast is a lightweight Personal News Radio CLI.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ezeoleaf/narrowcast/audio"
	"github.com/ezeoleaf/narrowcast/broadcast"
	"github.com/ezeoleaf/narrowcast/config"
	"github.com/ezeoleaf/narrowcast/state"
)

var version = "0.3.0"

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML or JSON config file")
	fetchTimeout := flag.Duration("fetch-timeout", 30*time.Second, "max time to wait for all feeds")
	tts := flag.String("tts", "", "override audio engine: mock, say, espeak, piper, elevenlabs, kokoro, auto")
	daemon := flag.Bool("daemon", false, "run forever on schedule.interval")
	once := flag.Bool("once", false, "single cycle; ignore schedule")
	stateFile := flag.String("state-file", "", "seen-URL JSON path (empty disables)")
	dryRun := flag.Bool("dry-run", false, "list matches without TTS")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("narrowcast", version)
		return
	}

	if err := run(*configPath, *fetchTimeout, *tts, *daemon, *once, *stateFile, *dryRun); err != nil {
		log.Fatalf("narrowcast: %v", err)
	}
}

func run(configPath string, fetchTimeout time.Duration, ttsOverride string, daemonFlag, onceFlag bool, stateFileFlag string, dryRun bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if ttsOverride != "" {
		cfg.Audio.Engine = ttsOverride
	}

	var speaker audio.Speaker
	if dryRun {
		speaker = nil
		log.Printf("dry-run: no TTS")
	} else {
		var desc string
		speaker, desc, err = audio.NewSpeaker(cfg.Audio)
		if err != nil {
			return err
		}
		log.Printf("tts: %s", desc)
	}

	interval, err := cfg.ScheduleInterval()
	if err != nil {
		return err
	}

	runDaemon := daemonFlag || (interval > 0 && !onceFlag)
	if onceFlag {
		runDaemon = false
	}

	statePath := resolveStatePath(cfg, runDaemon, onceFlag, stateFileFlag)
	ttl, err := cfg.StateTTLDuration()
	if err != nil {
		return err
	}
	store, err := state.Load(statePath, state.Options{
		TTL:        ttl,
		MaxEntries: cfg.State.MaxEntries,
	})
	if err != nil {
		return err
	}
	if statePath != "" {
		log.Printf("state file: %s", statePath)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runCycle := func() (broadcast.Result, error) {
		return broadcast.Cycle(ctx, cfg, speaker, store, fetchTimeout, dryRun)
	}

	if !runDaemon {
		return finishCycle(runCycle())
	}

	log.Printf("daemon mode: polling every %s", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := finishCycle(runCycle()); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func resolveStatePath(cfg *config.Config, daemon, once bool, flagPath string) string {
	if flagPath == "none" || flagPath == "off" {
		return ""
	}
	if flagPath != "" {
		return flagPath
	}
	if daemon {
		return cfg.State.File
	}
	if once && !cfg.State.Persist {
		return ""
	}
	if cfg.State.Persist || !once {
		return cfg.State.File
	}
	return ""
}

func finishCycle(res broadcast.Result, err error) error {
	if err != nil {
		return err
	}
	if res.Matched == 0 {
		fmt.Println("No matching headlines. Try broadening topics or keywords.")
	} else if res.Spoken == 0 && res.Skipped > 0 {
		fmt.Println("All matching headlines were already read. Clear the state file or wait for new articles.")
	} else if res.Queued > 0 {
		log.Printf("done: queued=%d spoken=%d skipped=%d", res.Queued, res.Spoken, res.Skipped)
	}
	return nil
}
