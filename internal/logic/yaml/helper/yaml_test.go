package helper

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestFindServicesNode(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected bool
	}{
		{
			name: "Valid services section",
			yaml: `
services:
  app:
    image: test
volumes:
  data: {}`,
			expected: true,
		},
		{
			name: "No services section",
			yaml: `
volumes:
  data: {}
networks:
  test:
    driver: bridge`,
			expected: false,
		},
		{
			name:     "Empty document",
			yaml:     `{}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yaml), &node)
			assert.NoError(t, err)

			result := FindServicesNode(&node)
			if tt.expected {
				assert.NotNil(t, result)
				assert.Equal(t, yaml.MappingNode, result.Kind)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
