package weiroll

// PlannerOption configures a Planner.
type PlannerOption func(*Planner)

// PlanOption configures the Plan() operation.
type PlanOption func(*planConfig)

// planConfig holds configuration for the Plan() method.
type planConfig struct {
	optimizeSlots bool
	maxCommands   int
	maxStateSlots int
}

// defaultPlanConfig returns the default plan configuration.
func defaultPlanConfig() *planConfig {
	return &planConfig{
		optimizeSlots: true,
		maxCommands:   256,
		maxStateSlots: MaxStateSlots,
	}
}

// WithSlotOptimization enables or disables aggressive slot reuse.
// When enabled (default), slots are recycled after their last usage.
func WithSlotOptimization(enabled bool) PlanOption {
	return func(c *planConfig) {
		c.optimizeSlots = enabled
	}
}

// WithMaxCommands sets a maximum command limit for the plan.
// Default is 256 commands.
func WithMaxCommands(max int) PlanOption {
	return func(c *planConfig) {
		c.maxCommands = max
	}
}

// WithMaxStateSlots sets a maximum state slot limit.
// Default is 127 (MaxStateSlots).
func WithMaxStateSlots(max int) PlanOption {
	return func(c *planConfig) {
		if max > MaxStateSlots {
			max = MaxStateSlots
		}
		c.maxStateSlots = max
	}
}
