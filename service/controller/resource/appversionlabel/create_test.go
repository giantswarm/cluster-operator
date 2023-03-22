package appversionlabel

import (
	"testing"
)

func Test_AppVersionLabelCreate(t *testing.T) {
	testCases := []struct {
		description      string
		currentVersion   string
		componentVersion string
		expectedOutcome  bool
	}{
		{
			description:      "Version is the same",
			currentVersion:   "5.11.0",
			componentVersion: "5.11.0",
			expectedOutcome:  false,
		},
		{
			description:      "Version should be updated (e.g. cluster was updated to a new release)",
			currentVersion:   "5.9.0",
			componentVersion: "5.11.0",
			expectedOutcome:  true,
		},
		{
			description:      "Version should not be updated (e.g. App CR is an Ap Bundle handled by the MC app-operator)",
			currentVersion:   "0.0.0",
			componentVersion: "5.11.0",
			expectedOutcome:  false,
		},
		{
			description:      "Should be able to hand over App CRs to the MC app-operator tho",
			currentVersion:   "5.9.0",
			componentVersion: "0.0.0",
			expectedOutcome:  true,
		},
		{
			description:      "Special case for same version",
			currentVersion:   "0.0.0",
			componentVersion: "0.0.0",
			expectedOutcome:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if shouldUpdateAppOperatorVersionLabel(tc.currentVersion, tc.componentVersion) != tc.expectedOutcome {
				t.Fatalf("Expected to update label: %v, but it would turn out the other way", tc.expectedOutcome)
			}
		})
	}
}
