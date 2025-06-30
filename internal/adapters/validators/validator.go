package validators

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func (v Validations) ToValidators() ([]Validator, error) {
	log.Debug().Int("count", len(v.Validations)).Msg("converting validations to validators")

	if len(v.Validations) == 0 {
		log.Info().Msg("No validations defined, skipping conversion to validators")
		return nil, nil
	}

	var result []Validator
	for i, v := range v.Validations {
		log.Debug().Int("index", i).Msg("processing validation")

		if v.Opa != nil {
			log.Debug().Int("index", i).Msg("creating OpaValidatorAdapter")
			validator := NewOpaValidatorAdapter(v.Opa, "opa-validator")
			result = append(result, validator)
		} else {
			log.Error().Int("index", i).Interface("validation", v).Msg("unknown validation type encountered")
			return nil, fmt.Errorf("unknown validation type: %v", v)
		}
	}

	log.Info().Int("count", len(result)).Msg("successfully converted validations to validators")
	return result, nil
}
