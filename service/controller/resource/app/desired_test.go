package app

import (
	"fmt"
	"testing"

	g8sv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
)

func Test_convertAndValidatePriority(t *testing.T) {
	testCases := []struct {
		description      string
		priorityStr      string
		expectedPriority int
		expectError      bool
	}{
		{
			description:      "case 1: not a number, fall back to default and return error",
			priorityStr:      "not-a-number",
			expectedPriority: g8sv1alpha1.ConfigPriorityDefault,
			expectError:      true,
		},
		{
			description:      "case 2: priority given is too low",
			priorityStr:      fmt.Sprintf("%d", g8sv1alpha1.ConfigPriorityCatalog-1),
			expectedPriority: g8sv1alpha1.ConfigPriorityDefault,
			expectError:      true,
		},
		{
			description:      "case 3: priority given is too high",
			priorityStr:      fmt.Sprintf("%d", g8sv1alpha1.ConfigPriorityMaximum+1),
			expectedPriority: g8sv1alpha1.ConfigPriorityDefault,
			expectError:      true,
		},
		{
			description:      "case 4: priority given is valid",
			priorityStr:      "42",
			expectedPriority: 42,
			expectError:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resultPriority, err := convertAndValidatePriority(tc.priorityStr)

			if resultPriority != tc.expectedPriority {
				t.Fatalf("Expected priority %d but got %d", tc.expectedPriority, resultPriority)
			}

			if err != nil && !tc.expectError {
				t.Fatalf("Got an unexpected error %#v", err)
			}
		})
	}
}
