package gemini

import (
	"context"
	"testing"
)

func TestListModels_ReturnsAllTiers(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 5 {
		t.Fatalf("got %d models, want 5", len(models))
	}

	tiers := make(map[string]bool)
	for _, m := range models {
		tiers[m.Tier] = true
	}
	for _, want := range []string{"nb2", "pro", "lite", "fast", "standard"} {
		if !tiers[want] {
			t.Errorf("missing tier %q in model list", want)
		}
	}
}

func TestListModels_MediaTypes(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	models, _ := p.ListModels(context.Background())

	imageTiers := 0
	videoTiers := 0
	for _, m := range models {
		switch m.MediaType {
		case "image":
			imageTiers++
		case "video":
			videoTiers++
		}
	}
	if imageTiers != 2 {
		t.Errorf("got %d image tiers, want 2", imageTiers)
	}
	if videoTiers != 3 {
		t.Errorf("got %d video tiers, want 3", videoTiers)
	}
}
