// Test package
package artifact_registry_test

import (
	pb "artifact-registry/protos"
	registry "artifact-registry/registry"
	"fmt"
)

// Usage guide to get artifacts by IDs
func ExampleMLArtifactStore_GetArtifactsByID() {
	artifactStore := registry.ArtifactStore("run-uuid")

	artifact := &pb.MLArtifact{
		// ArtifactType: pb.MLArtifact_MODEL,
		Ids: []int64{9474},
	}
	response, _ := artifactStore.GetArtifactsByID(artifact)

	for _, artifactData := range response.Artifacts {
		fmt.Println(artifactData.GetName())
		fmt.Println(artifactData.GetUri())
		fmt.Println(artifactData.GetVersion())
	}
	// Output:
	// FunctionComponent
	// example.com
	// FunctionComponent
}
