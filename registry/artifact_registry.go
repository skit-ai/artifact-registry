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

// Package artifact_registry provides interface to access the artifacts created
// by Kubeflow via MLMD.
//
// See Also
//
// <repository-readme-here>
package artifact_registry

import (
	pb "artifact-registry/protos"
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Vernacular-ai/vcore/log"
	"google.golang.org/grpc"
)

var (
	logLevel int
	client   pb.MetadataStoreServiceClient
)

// MLArtifactStore type provides access to list of Go methods to fetch
// artifacts in different ways.
type MLArtifactStore struct {
	Uuid string
	Host string
	Port string
}

// ArtifactStore function instantiates the MLArtifactStore instance.
//
// Use this instance to call methods to fetch artifacts, lineage tracking etc.
// Check examples/ for sample code snippets.
func ArtifactStore(uuid string) MLArtifactStore {
	var err error

	if logLevel, err = strconv.Atoi(strings.TrimSpace(os.Getenv("LOG_LEVEL"))); err == nil {
		log.SetLevel(logLevel)
	}

	artifactStore := MLArtifactStore{Uuid: uuid}

	client = clientInit()

	return artifactStore
}

// GetArtifactsByID - Fetch artifacts by list of artifact IDs
func (artifactStore MLArtifactStore) GetArtifactsByID(artifact *pb.MLArtifact) (pb.ArtifactsResponse, error) {
	var artifactsResponse pb.ArtifactsResponse

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	artifacts := &pb.GetArtifactsByIDRequest{
		ArtifactIds: artifact.Ids,
	}

	var err error
	response, err := client.GetArtifactsByID(ctx, artifacts)
	if err != nil {
		log.Debugf("Failed to fetch artifacts: %v", err)
		return artifactsResponse, err
	}

	var artifactList []*pb.ArtifactData
	for _, element := range response.Artifacts {
		artifactData := &pb.ArtifactData{
			Name:    element.Properties["name"].GetStringValue(),
			Uri:     element.GetUri(),
			Version: element.Properties["name"].GetStringValue(),
		}
		artifactList = append(artifactList, artifactData)
	}

	log.Debug("Fetch Artifacts by IDs")

	artifactsResponse = pb.ArtifactsResponse{Artifacts: artifactList}

	return artifactsResponse, nil
}

// Initialize gRPC client for MLMD storage connection
func clientInit() pb.MetadataStoreServiceClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("127.0.0.1:8080", opts...)
	if err != nil {
		log.Debugf("Failed to establish client connection: %v", err)
	}

	log.Debug("Client connected")

	client := pb.NewMetadataStoreServiceClient(conn)

	return client
}
