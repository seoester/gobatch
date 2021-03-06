package batch

import "time"

// Config retrieves the config values used by Batch. If these values are
// constant, NewConstantConfig can be used to create an implementation
// of the interface.
type Config interface {
	// Get returns the values for configuration.
	//
	// If MinItems > MaxItems or MinTime > MaxTime, the min value will be
	// set to the maximum value.
	//
	// If the config values may be modified during batch processing, Get
	// must properly handle concurrency issues.
	Get() ConfigValues
}

// ConfigValues is a struct that contains the Batch config values.
type ConfigValues struct {
	// MinTime specifies that a minimum amount of time that should pass
	// before processing items. The exception to this is if a max number
	// of items was specified and that number is reached before MinTime;
	// in that case those items will be processed right away.
	MinTime time.Duration `json:"minTime"`

	// MinItems specifies that a minimum number of items should be
	// processed at a time. Items will not be processed until MinItems
	// items are ready for processing. The exceptions to that are if MaxTime
	// is specified and that time is reached before the minimum number of
	// items is available, or if all items have been read and are ready
	// to process.
	MinItems uint64 `json:"minItems"`

	// MaxTime specifies that a maximum amount of time should pass before
	// processing. Once that time has been reached, items will be processed
	// whether or not MinItems items are available.
	MaxTime time.Duration `json:"maxTime"`

	// MaxItems specifies that a maximum number of items should be available
	// before processing. Once that number of items is available, they will
	// be processed whether or not MinTime has been reached.
	MaxItems uint64 `json:"maxItems"`
}

// NewConstantConfig returns a Config with constant values. If values
// is nil, the default values are used as described in Batch.
func NewConstantConfig(values *ConfigValues) *ConstantConfig {
	if values == nil {
		return &ConstantConfig{}
	}

	return &ConstantConfig{
		values: *values,
	}
}

// ConstantConfig is a Config with constant values. Create one with
// NewConstantConfig.
type ConstantConfig struct {
	values ConfigValues
}

// Get implements the Config interface.
func (b *ConstantConfig) Get() ConfigValues {
	return b.values
}
