package workingdir

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/rs/zerolog/log"
)

func listDirs(path string, nameRegex string) ([]string, error) {
	log.Debug().
		Str("nameRegex", nameRegex).
		Msg("compiling regex for directory names")
	regex, err := regexp.Compile(nameRegex)
	if err != nil {
		log.Error().
			Err(err).
			Str("nameRegex", nameRegex).
			Msg("failed to compile regex pattern")
		return nil, err
	}

	log.Debug().Str("path", path).Msg("reading directory entries")
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("failed to read directory")
		return nil, err
	}

	log.Debug().Int("totalEntries", len(entries)).Msg("processing directory entries")
	var dirs []string
	skippedFiles := 0
	regexFiltered := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			skippedFiles++
			continue
		}
		name := entry.Name()
		if regex != nil && !regex.MatchString(name) {
			log.Debug().
				Str("name", name).
				Str("regex", nameRegex).
				Msg("directory name does not match regex, skipping")
			regexFiltered++
			continue
		}
		fullPath := filepath.Join(path, name)
		log.Debug().
			Str("name", name).
			Msg("adding directory to results")
		dirs = append(dirs, fullPath)
	}

	log.Debug().
		Int("foundDirectories", len(dirs)).
		Int("skippedFiles", skippedFiles).
		Int("regexFiltered", regexFiltered).
		Str("path", path).
		Msg("completed directory listing")

	return dirs, nil
}
