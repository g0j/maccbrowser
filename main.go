package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/goj/maccbrowser/runner"
	"github.com/rs/zerolog"
)

func main() {
	logOutput := zerolog.ConsoleWriter{Out: os.Stdout}
	mainLogger := zerolog.New(logOutput).With().Timestamp().Logger()

	mainLogger.Info().Msg("starting")

	ctx, ctxCancel := context.WithCancel(context.Background())

	dockerRunner, err := runner.NewRunner(ctx)
	if err != nil {
		mainLogger.Fatal().Err(err).Msg("failed to create docker client")
	}

	signalChan := make(chan os.Signal, 1)

	mainLogger.Info().Msg("waiting for signal")
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signalChan

	mainLogger.Fatal().Str("signal", sig.String()).Msg("signal received")
	ctxCancel()
	dockerRunner.Close()
}
