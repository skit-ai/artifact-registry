/* Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
==============================================================================*/

// Utility functions

package artifact_registry

import (
	"context"
	"time"

	pb "artifact-registry/protos"
)

func prepareArtifactsList(artifacts []*pb.Artifact) []*pb.ArtifactData {
	artifactTypeMap := make(map[int]pb.ArtifactData_ArtifactType)

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	artifactTypes, _ := client.GetArtifactTypes(ctx, &pb.GetArtifactTypesRequest{})
	for _, artifactType := range artifactTypes.ArtifactTypes {
		switch artifactType.GetName() {
		case "kubeflow.org/alpha/data_set":
			artifactTypeMap[int(artifactType.GetId())] = pb.ArtifactData_DATASET
		case "kubeflow.org/alpha/metrics":
			artifactTypeMap[int(artifactType.GetId())] = pb.ArtifactData_METRICS
		case "kubeflow.org/alpha/model":
			artifactTypeMap[int(artifactType.GetId())] = pb.ArtifactData_MODEL
		default:
			artifactTypeMap[int(artifactType.GetId())] = pb.ArtifactData_OTHER
		}
	}

	var artifactList []*pb.ArtifactData
	for _, item := range artifacts {
		artifactData := &pb.ArtifactData{
			Name:         item.Properties["name"].GetStringValue(),
			Uri:          item.GetUri(),
			Version:      item.Properties["version"].GetStringValue(),
			ExecutionId:  item.CustomProperties["__kf_run__"].GetStringValue(),
			ArtifactType: artifactTypeMap[int(item.GetTypeId())],
		}
		artifactList = append(artifactList, artifactData)
	}
	return artifactList
}
