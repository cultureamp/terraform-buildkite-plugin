package workingdir

import "github.com/rs/zerolog/log"

// Partition returns the contiguous chunk of items assigned to job `i`
// out of `n` total jobs. Jobs and jobCount are 0-based: 0 ≤ i < n.
// If jobCount ≤ 0 or i is out of range, it returns nil.
//
// It distributes len(items) as evenly as possible: the first
// (len(items)%n) chunks get one extra element.
func partition[T any](items []T, jobIndex, jobCount int) []T {
	length := len(items)
	log.Debug().
		Int("itemCount", length).
		Int("jobIndex", jobIndex).
		Int("jobCount", jobCount).
		Msg("partitioning items for parallel jobs")

	if jobCount <= 0 || jobIndex < 0 || jobIndex >= jobCount {
		log.Debug().Msg("invalid job parameters, returning nil")
		return nil
	}

	baseSize := length / jobCount
	extra := length % jobCount

	// compute start
	var start int
	if jobIndex < extra {
		start = jobIndex * (baseSize + 1)
	} else {
		start = extra*(baseSize+1) + (jobIndex-extra)*baseSize
	}

	// compute size
	var sz int
	if jobIndex < extra {
		sz = baseSize + 1
	} else {
		sz = baseSize
	}

	log.Debug().
		Int("start", start).
		Int("size", sz).
		Int("baseSize", baseSize).
		Int("extra", extra).
		Msg("computed partition parameters")

	result := items[start : start+sz]
	log.Debug().Int("resultCount", len(result)).Msg("partition completed")
	return result
}
