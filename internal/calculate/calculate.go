package calculate

import (
	"github.com/antpas14/fantalegheEV-api"
)

// GetRanks retrieves a list of ranks (api.Rank)
func GetRanks() []api.Rank {
	// Perform data retrieval and processing here
	// For this example, we'll return a static list of ranks
	return []api.Rank{
		{
			EvPoints: float64Ptr(float64(100.5)),
			Points:   intPtr(200),
			Team:     strPtr("Team A"),
		},
		{
			EvPoints: float64Ptr(float64(100.5)),
			Points:   intPtr(200),
			Team:     strPtr("Team A"),
		},
		// Add more ranks as needed...
	}
}

// Helper function to create a float64 pointer
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create an int pointer
func intPtr(i int) *int {
	return &i
}

// Helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}