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
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Vernacular-ai/vcore/log"
	"google.golang.org/grpc"

	pb "artifact-registry/protos"
)

var CONTEXT_TYPE_NAME string = "kubeflow.org/alpha/workspace"

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

// Workspace type provides access to list of Go methods to fetch artifacts
// grouped within a "workspace"
type Workspace struct {
    Id int64
    Name string
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
func (artifactStore MLArtifactStore) GetArtifactsByID(artifact *pb.MLArtifact) (*pb.ArtifactsResponse, error) {
	var artifactsResponse *pb.ArtifactsResponse

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

    artifactList := prepareArtifactsList(response.Artifacts)

	artifactsResponse = &pb.ArtifactsResponse{Artifacts: artifactList}

	return artifactsResponse, nil
}

// GetWorkspace - returns a workspace instance for the requested named
// workspace. This is used to call methods to fetch artifacts grouped in a
// particular workspace
func (artifactStore MLArtifactStore) GetWorkspace(workspace *pb.Workspace) (Workspace, error) {
    var workspaceResponse Workspace

    ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

    contextRequest := &pb.GetContextByTypeAndNameRequest{
        TypeName: &CONTEXT_TYPE_NAME,
        ContextName: &workspace.Name,
    }

    var err error
    response, err := client.GetContextByTypeAndName(ctx, contextRequest)
    if err != nil {
		log.Debugf("Failed to fetch workspace: %v", err)
		return workspaceResponse, err
	}

    workspaceResponse = Workspace{Id: response.Context.GetId(), Name: response.Context.GetName()}
	log.Debugf("Fetched workspace %s", response.Context.GetName())

    return workspaceResponse, nil
}

// GetArtifactsByWorkspace - returns a list of artifacts associated with this
// workspace
func (workspace Workspace) GetArtifactsByWorkspace() (*pb.ArtifactsResponse, error) {
	var artifactsResponse *pb.ArtifactsResponse

    contextRequest := &pb.GetArtifactsByContextRequest{ContextId: &workspace.Id}

    ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

    var err error
    response, err := client.GetArtifactsByContext(ctx, contextRequest)
    if err != nil {
        log.Debugf("Failed to fetch artifacts for workspace %s, Error: %v", workspace.Name, err)
        return artifactsResponse, err
    }

    artifactList := prepareArtifactsList(response.GetArtifacts())

	artifactsResponse = &pb.ArtifactsResponse{Artifacts: artifactList}

	return artifactsResponse, nil
}

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
