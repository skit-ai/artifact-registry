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
// https://github.com/Vernacular-ai/artifact-registry
package artifact_registry

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Vernacular-ai/vcore/log"
	"google.golang.org/grpc"

	pb "github.com/Vernacular-ai/artifact-registry/protos"
)

// Kubeflow default context
var CONTEXT_TYPE_NAME string = "kubeflow.org/alpha/workspace"

// Kubeflow type name for model type artifacts
var MODEL_ARTIFACT_TYPE_NAME = "kubeflow.org/alpha/model"

// Kubeflow type name for dataset type artifacts
var DATASET_ARTIFACT_TYPE_NAME = "kubeflow.org/alpha/data_set"

// Kubeflow type name for metrics type artifacts
var METRICS_ARTIFACT_TYPE_NAME = "kubeflow.org/alpha/metrics"

// Kubeflow type name for executions
var EXECUTION_TYPE_NAME = "kubeflow.org/alpha/execution"

var (
	logLevel int
	client   pb.MetadataStoreServiceClient
)

// MLArtifactStore type provides access to list of Go methods to fetch
// artifacts in different ways.
type MLArtifactStore struct {
	Host string
	Port string
}

// Workspace type provides access to list of Go methods to fetch artifacts
// grouped within a "workspace"
type Workspace struct {
	Id   int64
	Name string
}

// ArtifactStore function instantiates the MLArtifactStore instance.
//
// Use this instance to call methods to fetch artifacts, lineage tracking etc.
func ArtifactStore(host string, port string) MLArtifactStore {
	var err error

	if logLevel, err = strconv.Atoi(strings.TrimSpace(os.Getenv("LOG_LEVEL"))); err == nil {
		log.SetLevel(logLevel)
	}

	artifactStore := MLArtifactStore{Host: host, Port: port}
	client = clientInit(artifactStore)

	return artifactStore
}

// GetArtifactsByID fetches artifacts by list of artifact IDs
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

// GetWorkspace returns a workspace instance for the requested named
// workspace. This is used to call methods to fetch artifacts grouped in a
// particular workspace.
func (artifactStore MLArtifactStore) GetWorkspace(workspace *pb.Workspace) (Workspace, error) {
	var workspaceResponse Workspace

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	contextRequest := &pb.GetContextByTypeAndNameRequest{
		TypeName:    &CONTEXT_TYPE_NAME,
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

// GetArtifactsByWorkspace returns a list of artifacts associated with this
// workspace.
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

// GetArtifactsByTypeWorkspace returns a list of artifacts of a certain type
// associated with this workspace
func (workspace Workspace) GetArtifactsByTypeWorkspace(artifactTypeRequest *pb.ArtifactByTypeRequest) (*pb.ArtifactsResponse, error) {
	var artifactsResponse *pb.ArtifactsResponse
	var artifactType string

	switch artifactTypeRequest.ArtifactType {
	case pb.ArtifactByTypeRequest_DATASET:
		artifactType = DATASET_ARTIFACT_TYPE_NAME
	case pb.ArtifactByTypeRequest_MODEL:
		artifactType = MODEL_ARTIFACT_TYPE_NAME
	case pb.ArtifactByTypeRequest_METRICS:
		artifactType = METRICS_ARTIFACT_TYPE_NAME
	default:
		var err error
		log.Debugf("Artifact type %s does not exist", artifactTypeRequest.ArtifactType.Enum())
		return artifactsResponse, err
	}

	artifactsByTypeRequest := &pb.GetArtifactsByTypeRequest{TypeName: &artifactType}

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	response, err := client.GetArtifactsByType(ctx, artifactsByTypeRequest)
	if err != nil {
		log.Debugf("Failed to fetch artifacts for workspace %s, Error: %v", workspace.Name, err)
		return artifactsResponse, err
	}

	artifactList := prepareFilteredArtifactsList(response.GetArtifacts(), workspace.Name)

	artifactsResponse = &pb.ArtifactsResponse{Artifacts: artifactList}

	return artifactsResponse, nil
}

// GetLineageByRun returns a list of artifacts associated with a Kubeflow run
func (workspace Workspace) GetLineageByRun(artifactsByRunRequest *pb.ArtifactsByRunRequest) (*pb.ArtifactsResponse, error) {
	var artifactsResponse *pb.ArtifactsResponse
	var artifactList []*pb.ArtifactData

	workspaceArtifacts, _ := workspace.GetArtifactsByWorkspace()
	log.Debug(workspaceArtifacts)

	for _, artifactData := range workspaceArtifacts.GetArtifacts() {
		log.Debug(artifactData.GetId())
		if artifactData.GetRunId() == artifactsByRunRequest.GetRunId() {
			artifactList = append(artifactList, artifactData)
		}
	}

	artifactsResponse = &pb.ArtifactsResponse{Artifacts: artifactList}

	return artifactsResponse, nil
}

// GetLinageByModel returns a list of artifacts associated with a Model
func (workspace Workspace) GetLineageByModel(artifactsByModelRequest *pb.ArtifactsByModelRequest) (*pb.ArtifactsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	// All executions associated with this model
	eventsByArtifactIdRequest := &pb.GetEventsByArtifactIDsRequest{
		ArtifactIds: []int64{artifactsByModelRequest.GetModelId()},
	}

	response, _ := client.GetEventsByArtifactIDs(ctx, eventsByArtifactIdRequest)

	var executionIds []int64
	for _, event := range response.GetEvents() {
		executionIds = append(executionIds, event.GetExecutionId())
	}

	// All events associated with all the executions of this model
	eventsByExecutionIdsRequest := &pb.GetEventsByExecutionIDsRequest{
		ExecutionIds: uniqueList(executionIds),
	}

	responseEvents, _ := client.GetEventsByExecutionIDs(ctx, eventsByExecutionIdsRequest)

	// All the artifacts of the events
	var artifactIds []int64
	for _, event := range responseEvents.GetEvents() {
		artifactIds = append(artifactIds, event.GetArtifactId())
	}

	artifactsByIdsRequest := &pb.MLArtifact{
		Ids: uniqueList(artifactIds),
	}

	artifactStore := MLArtifactStore{}
	artifactsResponse, _ := artifactStore.GetArtifactsByID(artifactsByIdsRequest)

	return artifactsResponse, nil
}

func clientInit(artifactStore MLArtifactStore) pb.MetadataStoreServiceClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	address := fmt.Sprintf("%s:%s", artifactStore.Host, artifactStore.Port)
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Debugf("Failed to establish client connection: %v", err)
	}

	log.Debug("Client connected")

	client := pb.NewMetadataStoreServiceClient(conn)

	return client
}
