# artifact-registry

Go SDK to manage artifacts stored by [Kubeflow][kubeflow] via [MLMD][mlmd].

## Documentation

```
godoc -http:6060
```

Go to `http://localhost:6060/pkg/artifact-registry/registry/` to checkout the doc.

## Development

### Pre-requisites

#### 1. Using local MLMD deployment

Follow the doc here: https://github.com/google/ml-metadata/ to setup local MLMD gRPC server.

#### 2. Kubernetes deployed MLMD server

To access the MLMD gRPC server bundled with Kubeflow:

```
kubectl port-forward -n kubeflow $(kubectl get pods -nkubeflow | grep metadata-grpc-deployment | head -n 1 | cut -d' ' -f1) 8080:8080
```

### SDK

    .
    ├── LICENSE
    ├── README.md
    ├── go.mod
    ├── go.sum
    ├── protos
    │   ├── artifact-registry.pb.go
    │   ├── artifact-registry.proto
    │   ├── metadata_source.pb.go
    │   ├── metadata_source.proto
    │   ├── metadata_store.pb.go
    │   ├── metadata_store.proto
    │   ├── metadata_store_service.pb.go
    │   └── metadata_store_service.proto
    └── registry
        ├── artifact_registry.go
        ├── registry_test.go
        └── utils.go

- `protos/` has protobufs and generated code for MLMD data store, MLMD gRPC
  service and the artifact registry SDK's data definition.
- `registry/` has all the code to manage artifacts.


[kubeflow]: https://www.kubeflow.org/docs/about/kubeflow/
[mlmd]: https://github.com/google/ml-metadata/
