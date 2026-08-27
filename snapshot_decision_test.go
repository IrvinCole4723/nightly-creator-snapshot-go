package main

import "testing"

func TestShouldSnapshotWaitsForProcessedContent(t *testing.T) {
	input := CreatorSnapshot{CreatorID: "studio-17", Subscribers: 42, NewContent: 3, ProcessedAssets: 2}
	if shouldSnapshot(input) {
		t.Fatal("snapshot must wait for content processing")
	}
	input.ProcessedAssets = 3
	if !shouldSnapshot(input) {
		t.Fatal("snapshot should publish when subscribers and processed content are present")
	}
}
