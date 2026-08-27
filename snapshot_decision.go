package main

type CreatorSnapshot struct {
	CreatorID       string `json:"creator_id"`
	Subscribers     int    `json:"subscribers"`
	NewContent      int    `json:"new_content"`
	ProcessedAssets int    `json:"processed_assets"`
}

func shouldSnapshot(input CreatorSnapshot) bool {
	return input.CreatorID != "" && input.Subscribers > 0 && input.NewContent > 0 && input.ProcessedAssets == input.NewContent
}
