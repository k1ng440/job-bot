package utils

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func Mkdir(dir string) (string, func()) {
	var err error
	if dir == "" {
		dir, err = os.MkdirTemp("", "jb")
		if err != nil {
			log.Error().Err(err).Msg("Failed to create temp dir")
			os.Exit(1)
		}
		return dir, func() {
			os.RemoveAll(dir)
		}
	} else {
		dir, err = filepath.Abs(dir)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get absolute path")
			os.Exit(1)
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create dir")
			os.Exit(1)
		}

		// Do not remove the directory if it was provided by the user
		return dir, func() {}
	}
}
