package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/goj/maccbrowser/runner"
	"github.com/goj/maccbrowser/storage"
	"github.com/goj/maccbrowser/storage/disk"
	"github.com/rs/zerolog"
)

const profileRoot = "./profile_storage"

func main() {
	logOutput := zerolog.ConsoleWriter{Out: os.Stdout}
	mainLogger := zerolog.New(logOutput).With().Timestamp().Logger()

	mainLogger.Info().Msg("starting")

	ctx, ctxCancel := context.WithCancel(context.Background())

	dockerRunner, err := runner.NewRunner(ctx)
	if err != nil {
		mainLogger.Fatal().Err(err).Msg("failed to create docker client")
	}

	var profStor storage.ProfileStorage

	profStor, err = disk.NewStorage(profileRoot)
	if err != nil {
		mainLogger.Fatal().Err(err).Msg("failed to create profile storage")
	}

	err = profStor.Load()
	if err != nil {
		mainLogger.Fatal().Err(err).Msg("failed to init profile storage")
	}

	err = dockerRunner.RunBrowser(runner.RunOptions{
		ChromeVersion: "90.0",
		GUID:          "firstTryContainer",
	})
	if err != nil {
		mainLogger.Fatal().Err(err).Msg("failed to run browser")
	}

	signalChan := make(chan os.Signal, 1)

	mainLogger.Info().Msg("waiting for signal")
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signalChan

	mainLogger.Warn().Str("signal", sig.String()).Msg("signal received, exiting...")
	ctxCancel()

	err = dockerRunner.Close()
	if err != nil {
		mainLogger.Error().Err(err).Msg("failed to stop containers")
	}
}
