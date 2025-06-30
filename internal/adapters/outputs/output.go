package outputs

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func (o Outputs) ToOutputers() ([]Outputer, error) {
	if len(o.Outputs) == 0 {
		log.Info().Msg("No outputs defined, skipping conversion to outputers")
		return nil, nil
	}
	log.Debug().Int("count", len(o.Outputs)).Msg("converting outputs to outputers")
	var result []Outputer
	for i, o := range o.Outputs {
		log.Debug().Int("index", i).Msg("processing output")
		if o.BuildkiteAnnotation != nil {
			log.Debug().Int("index", i).Msg("creating BuildkiteAnnotator")
			output := NewBuildkiteAnnotator(WithConfig(o.BuildkiteAnnotation))
			result = append(result, output)
		} else {
			log.Error().Int("index", i).Interface("output", o).Msg("unknown output type encountered")
			return nil, fmt.Errorf("unknown output type: %v", o)
		}
	}
	log.Info().Int("count", len(result)).Msg("successfully converted outputs to outputers")
	return result, nil
}
