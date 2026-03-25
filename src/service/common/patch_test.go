package common

import (
	"testing"
)

func TestAssignStringIfPresent(t *testing.T) {
	tests := []struct {
		name      string
		updates   map[string]interface{}
		patch     map[string]interface{}
		key       string
		condition bool
		wantPatch map[string]interface{}
		wantErr   string
	}{
		{
			name:      "missing in updates",
			updates:   map[string]interface{}{"other": "value"},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{},
			wantErr:   "",
		},
		{
			name:      "present but condition false",
			updates:   map[string]interface{}{"mykey": "value"},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: false,
			wantPatch: map[string]interface{}{},
			wantErr:   "cannot edit mykey field",
		},
		{
			name:      "present and empty string",
			updates:   map[string]interface{}{"mykey": ""},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{"mykey": nil},
			wantErr:   "",
		},
		{
			name:      "present and valid string",
			updates:   map[string]interface{}{"mykey": "new_value"},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{"mykey": "new_value"},
			wantErr:   "",
		},
		{
			name:      "present but not a string type",
			updates:   map[string]interface{}{"mykey": 123},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{}, // type assertion fails, silently ignores
			wantErr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssignStringIfPresent(tt.updates, tt.patch, tt.key, tt.condition)
			if err != nil {
				if tt.wantErr == "" {
					t.Errorf("AssignStringIfPresent() unexpected error: %v", err)
				} else {
					if err.Error() != tt.wantErr {
						t.Errorf("AssignStringIfPresent() error = %v, wantErr %v", err.Error(), tt.wantErr)
					}
				}
			} else if tt.wantErr != "" {
				t.Errorf("AssignStringIfPresent() expected error: %v but got none", tt.wantErr)
			}

			// check patch contents
			if len(tt.patch) != len(tt.wantPatch) {
				t.Errorf("AssignStringIfPresent() patch len = %v, want %v", len(tt.patch), len(tt.wantPatch))
			} else {
				for k, v := range tt.wantPatch {
					if tt.patch[k] != v {
						t.Errorf("AssignStringIfPresent() patch[%s] = %v, want %v", k, tt.patch[k], v)
					}
				}
			}
		})
	}
}

func TestAssignMandatoryString(t *testing.T) {
	tests := []struct {
		name      string
		updates   map[string]interface{}
		patch     map[string]interface{}
		key       string
		condition bool
		wantPatch map[string]interface{}
		wantErr   string
	}{
		{
			name:      "missing in updates",
			updates:   map[string]interface{}{"other": "value"},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{},
			wantErr:   "",
		},
		{
			name:      "present but condition false",
			updates:   map[string]interface{}{"mykey": "value"},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: false,
			wantPatch: map[string]interface{}{},
			wantErr:   "cannot edit mykey field",
		},
		{
			name:      "present and empty string",
			updates:   map[string]interface{}{"mykey": ""},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{},
			wantErr:   "mykey is mandatory",
		},
		{
			name:      "present and valid string",
			updates:   map[string]interface{}{"mykey": "new_value"},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{"mykey": "new_value"},
			wantErr:   "",
		},
		{
			name:      "present but not a string type",
			updates:   map[string]interface{}{"mykey": 123},
			patch:     map[string]interface{}{},
			key:       "mykey",
			condition: true,
			wantPatch: map[string]interface{}{},
			wantErr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssignMandatoryString(tt.updates, tt.patch, tt.key, tt.condition)
			if err != nil {
				if tt.wantErr == "" {
					t.Errorf("AssignMandatoryString() unexpected error: %v", err)
				} else {
					if err.Error() != tt.wantErr {
						t.Errorf("AssignMandatoryString() error = %v, wantErr %v", err.Error(), tt.wantErr)
					}
				}
			} else if tt.wantErr != "" {
				t.Errorf("AssignMandatoryString() expected error: %v but got none", tt.wantErr)
			}

			// check patch contents
			if len(tt.patch) != len(tt.wantPatch) {
				t.Errorf("AssignMandatoryString() patch len = %v, want %v", len(tt.patch), len(tt.wantPatch))
			} else {
				for k, v := range tt.wantPatch {
					if tt.patch[k] != v {
						t.Errorf("AssignMandatoryString() patch[%s] = %v, want %v", k, tt.patch[k], v)
					}
				}
			}
		})
	}
}
