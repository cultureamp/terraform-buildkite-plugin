package workingdir

import (
	"errors"

	"github.com/rs/zerolog/log"
)

func (w *Working) Parse() ([]string, error) {
	log.Debug().Msg("parsing working directory configuration")

	if w == nil {
		log.Debug().Msg("working directory configuration is nil, returning empty slice")
		// TODO: Confirm this behavior is acceptable, or if we should return an error
		// return nil, errors.New("plugin configuration is nil")
		return []string{}, nil // Return empty slice instead of error for missing config
	}

	// If the plugin has a single working directory, return it directly
	if w.Directory != nil && *w.Directory != "" {
		log.Debug().Str("directory", *w.Directory).Msg("using single working directory")
		return []string{*w.Directory}, nil
	}

	if w.Directories != nil {
		log.Debug().Msg("processing multiple working directories")
		directories, err := handleWorkingDirectories(w.Directories)
		if err != nil {
			log.Error().Err(err).Msg("failed to handle working directories")
			return nil, err
		}

		// if the plugin has parallelism configured, partition the directories
		if w.Parallelism != nil {
			result := partition(directories, *w.Parallelism.ParallelJob, *w.Parallelism.ParallelJobCount)
			log.Info().
				Int("parallelJob", *w.Parallelism.ParallelJob).
				Int("parallelJobCount", *w.Parallelism.ParallelJobCount).
				Int("selectedDirectoryCount", len(result)).
				Int("totalDirectoryCount", len(directories)).
				Interface("directories", result).
				Msg("successfully parsed working directories with parallelism")
			return result, nil
		}

		// if no parallelism is configured, return the directories as is
		log.Info().
			Int("count", len(directories)).
			Interface("directories", directories).
			Msg("successfully parsed working directories")
		return directories, nil
	}

	log.Error().Msg("no valid working directory configuration found")
	return nil, errors.New("no valid working directory configuration found")
}

func handleWorkingDirectories(w *Directories) ([]string, error) {
	log.Debug().Msg("handling working directories configuration")

	if w == nil {
		log.Error().Msg("working directories configuration is nil")
		return nil, errors.New("working directories configuration is nil")
	}

	if w.ParentDirectory != "" {
		c, err := listDirs(w.ParentDirectory, w.NameRegex)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	if w.Artifact != "" {
		// TODO implement the logic to handle artifacts
		log.Debug().Str("artifact", w.Artifact).Msg("processing artifact configuration")
		// We should download the artifact then extract it to a temporary directory then apply the name regex to find matching directories
		log.Warn().Str("artifact", w.Artifact).Msg("Artifact handling not implemented yet")
		return nil, errors.New("artifact handling not implemented yet")
	}

	log.Error().Msg("no valid working directory configuration found in directories config")
	return nil, errors.New("no valid working directory configuration found")
}
