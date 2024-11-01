// Package helper provides utility functions for working with YAML nodes
package helper

import "gopkg.in/yaml.v3"

// FindServicesNode locates the services section in the YAML tree
// Parameters:
//   - node: The root YAML node to search in
//
// Returns:
//   - *yaml.Node: The services section node if found, nil otherwise
func FindServicesNode(node *yaml.Node) *yaml.Node {
	if node.Kind != yaml.DocumentNode {
		return nil
	}

	rootMap := node.Content[0]
	for i := 0; i < len(rootMap.Content); i += 2 {
		if rootMap.Content[i].Value == "services" {
			return rootMap.Content[i+1]
		}
	}
	return nil
}
