package artifact_registry

import (
	pb "artifact-registry/protos"
)


func prepareArtifactsList(artifacts []*pb.Artifact) []*pb.ArtifactData {
	var artifactList []*pb.ArtifactData
	for _, element := range artifacts {
		artifactData := &pb.ArtifactData{
			Name:    element.Properties["name"].GetStringValue(),
			Uri:     element.GetUri(),
			Version: element.Properties["name"].GetStringValue(),
		}
		artifactList = append(artifactList, artifactData)
	}
    return artifactList
}
