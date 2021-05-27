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

// Example usage to find Workspace by workspace name
func ExampleMLArtifactStore_GetWorkspace() {
	artifactStore := registry.ArtifactStore("run-uuid")

    workspace := &pb.Workspace{
        Name: "workspace_1",
    }

    response, _ := artifactStore.GetWorkspace(workspace)

    fmt.Println(response.Name)
    // Output:
    // workspace_1
}

// Example to fetch artifacts in a workspace
func ExampleWorkspace_GetArtifactsByWorkspace() {
	artifactStore := registry.ArtifactStore("run-uuid")

    workspaceInfo := &pb.Workspace{
        Name: "workspace_1",
    }

    workspace, _ := artifactStore.GetWorkspace(workspaceInfo)
    fmt.Println(workspace.Name)

    artifactList, _ := workspace.GetArtifactsByWorkspace()

    for _, artifactData := range(artifactList.GetArtifacts()) {
        fmt.Println(artifactData.GetName())
    }
}

