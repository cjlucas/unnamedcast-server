package main

import (
	"testing"

	"github.com/cjlucas/unnamedcast/worker/api"
)

func TestUpdateFeedWorker_mergeFeeds(t *testing.T) {
	f1 := api.Feed{
		Items: []api.Item{
			{GUID: "1", Title: "1"},
			{GUID: "2", Title: "2"},
			{GUID: "3", Title: "3"},
		},
	}

	f2 := api.Feed{
		Items: []api.Item{
			{GUID: "3", Title: "4"},
			{GUID: "4", Title: "5"},
			{GUID: "5", Title: "6"},
		},
	}

	w := UpdateFeedWorker{}
	f := w.mergeFeeds(&f1, &f2)
	if len(f.Items) != 5 {
		t.Errorf("Resulting feed should have 5 items, has %d", len(f.Items))
	}

	guidMap := make(map[string]*api.Item)
	for i := range f.Items {
		guidMap[f.Items[i].GUID] = &f.Items[i]
	}

	for _, s := range []string{"1", "2", "3", "4", "5"} {
		if _, ok := guidMap[s]; !ok {
			t.Errorf("Expected item with GUID \"%s\" to exist", s)
		}
	}

	if guidMap["3"].Title != "4" {
		t.Errorf("Expected duplicate item to overrwrite with new information")
	}
}
