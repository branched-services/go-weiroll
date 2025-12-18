package weiroll

import (
	"testing"
)

func TestDefaultPlanConfig(t *testing.T) {
	config := defaultPlanConfig()

	t.Run("slot optimization enabled by default", func(t *testing.T) {
		if !config.optimizeSlots {
			t.Error("Expected optimizeSlots to be true by default")
		}
	})

	t.Run("max commands is 256 by default", func(t *testing.T) {
		if config.maxCommands != 256 {
			t.Errorf("Expected maxCommands to be 256, got %d", config.maxCommands)
		}
	})

	t.Run("max state slots is MaxStateSlots by default", func(t *testing.T) {
		if config.maxStateSlots != MaxStateSlots {
			t.Errorf("Expected maxStateSlots to be %d, got %d", MaxStateSlots, config.maxStateSlots)
		}
	})
}

func TestWithSlotOptimization(t *testing.T) {
	t.Run("disables slot optimization", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithSlotOptimization(false)
		opt(config)

		if config.optimizeSlots {
			t.Error("Expected optimizeSlots to be false")
		}
	})

	t.Run("enables slot optimization", func(t *testing.T) {
		config := defaultPlanConfig()
		config.optimizeSlots = false // Start disabled
		opt := WithSlotOptimization(true)
		opt(config)

		if !config.optimizeSlots {
			t.Error("Expected optimizeSlots to be true")
		}
	})
}

func TestWithMaxCommands(t *testing.T) {
	t.Run("sets custom max commands", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxCommands(100)
		opt(config)

		if config.maxCommands != 100 {
			t.Errorf("Expected maxCommands to be 100, got %d", config.maxCommands)
		}
	})

	t.Run("allows setting higher than default", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxCommands(1000)
		opt(config)

		if config.maxCommands != 1000 {
			t.Errorf("Expected maxCommands to be 1000, got %d", config.maxCommands)
		}
	})

	t.Run("allows setting to zero", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxCommands(0)
		opt(config)

		if config.maxCommands != 0 {
			t.Errorf("Expected maxCommands to be 0, got %d", config.maxCommands)
		}
	})
}

func TestWithMaxStateSlots(t *testing.T) {
	t.Run("sets custom max state slots", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxStateSlots(50)
		opt(config)

		if config.maxStateSlots != 50 {
			t.Errorf("Expected maxStateSlots to be 50, got %d", config.maxStateSlots)
		}
	})

	t.Run("caps at MaxStateSlots", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxStateSlots(500) // Greater than MaxStateSlots (127)
		opt(config)

		if config.maxStateSlots != MaxStateSlots {
			t.Errorf("Expected maxStateSlots to be capped at %d, got %d", MaxStateSlots, config.maxStateSlots)
		}
	})

	t.Run("allows setting exactly MaxStateSlots", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxStateSlots(MaxStateSlots)
		opt(config)

		if config.maxStateSlots != MaxStateSlots {
			t.Errorf("Expected maxStateSlots to be %d, got %d", MaxStateSlots, config.maxStateSlots)
		}
	})

	t.Run("allows setting to zero", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxStateSlots(0)
		opt(config)

		if config.maxStateSlots != 0 {
			t.Errorf("Expected maxStateSlots to be 0, got %d", config.maxStateSlots)
		}
	})

	t.Run("allows setting to one", func(t *testing.T) {
		config := defaultPlanConfig()
		opt := WithMaxStateSlots(1)
		opt(config)

		if config.maxStateSlots != 1 {
			t.Errorf("Expected maxStateSlots to be 1, got %d", config.maxStateSlots)
		}
	})
}

func TestMultipleOptions(t *testing.T) {
	config := defaultPlanConfig()

	opts := []PlanOption{
		WithSlotOptimization(false),
		WithMaxCommands(50),
		WithMaxStateSlots(25),
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.optimizeSlots {
		t.Error("Expected optimizeSlots to be false")
	}
	if config.maxCommands != 50 {
		t.Errorf("Expected maxCommands to be 50, got %d", config.maxCommands)
	}
	if config.maxStateSlots != 25 {
		t.Errorf("Expected maxStateSlots to be 25, got %d", config.maxStateSlots)
	}
}

func TestOptionsOrderMatters(t *testing.T) {
	t.Run("last option wins", func(t *testing.T) {
		config := defaultPlanConfig()

		opts := []PlanOption{
			WithMaxCommands(100),
			WithMaxCommands(200),
			WithMaxCommands(300),
		}

		for _, opt := range opts {
			opt(config)
		}

		if config.maxCommands != 300 {
			t.Errorf("Expected last value (300), got %d", config.maxCommands)
		}
	})
}

func TestPlanConfigIndependence(t *testing.T) {
	t.Run("each config is independent", func(t *testing.T) {
		config1 := defaultPlanConfig()
		config2 := defaultPlanConfig()

		opt := WithMaxCommands(99)
		opt(config1)

		if config1.maxCommands != 99 {
			t.Errorf("Expected config1.maxCommands to be 99, got %d", config1.maxCommands)
		}
		if config2.maxCommands != 256 {
			t.Errorf("Expected config2.maxCommands to remain 256, got %d", config2.maxCommands)
		}
	})
}

func TestPlannerOptionType(t *testing.T) {
	// PlannerOption is a function that takes *Planner
	// This test just verifies the type exists and is usable
	var _ PlannerOption = func(p *Planner) {
		// Example planner option
	}
}

func TestPlanOptionType(t *testing.T) {
	// PlanOption is a function that takes *planConfig
	// This test verifies the type matches our options
	var opts []PlanOption
	opts = append(opts, WithSlotOptimization(true))
	opts = append(opts, WithMaxCommands(100))
	opts = append(opts, WithMaxStateSlots(50))

	if len(opts) != 3 {
		t.Errorf("Expected 3 options, got %d", len(opts))
	}
}
