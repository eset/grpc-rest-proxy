# Helm chart for grpc-rest-propxy

- It is recommended to run grpc-rest-proxy as sidecar proxy (see [../../README.md] for more info).
- Helm chart can be used to run grpc-rest-proxy as standalone Kubernetes depoloyment.

## Example usage

```bash
helm upgrade \
    --install \
    --namespace ${GRPC_REST_PROXY_NAMESPACE} \
    --set grpcRestProxy.grpcTargetAddr=${GRPC_REST_PROXY_TARGET_GRPC_SERVER_ADDR} \
    --image.repository=${GRPC_REST_PROXY_CONTAINER_REPOSITORY}/grpc-rest-proxy \
    --image.tag=latest \
    grpc-rest-proxy ./charts/grpc-rest-proxy
```

### Example usage with example-grpc-server

#### 1. Prepare app containers

Build and push containers for grpc-rest-proxy and example-grpc-server

```bash
docker build -f cmd/examples/grpcserver/Dockerfile -t ${GRPC_REST_PROXY_CONTAINER_REPOSITORY}/example-grpc-server:latest .
docker build -f Dockerfile -t ${GRPC_REST_PROXY_CONTAINER_REPOSITORY}/grpc-rest-proxy:latest .

docker push ${GRPC_REST_PROXY_CONTAINER_REPOSITORY}/example-grpc-server:latest
docker push ${GRPC_REST_PROXY_CONTAINER_REPOSITORY}/grpc-rest-proxy:latest
```

#### 2. Deploy example-grpc-server as standalone Kubernetes deployment

Update `spec.template.spec.containers[0].image` with existing container repository

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: example-grpc-server
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-grpc-server
  namespace: example-grpc-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example-grpc-server
  template:
    metadata:
      labels:
        app: example-grpc-server
    spec:
      containers:
      - name: example-grpc-server
        image: ${GRPC_REST_PROXY_CONTAINER_REPOSITORY}/example-grpc-server:latest
        ports:
        - containerPort: 50051
        command: ["/app/example-grpc-server"]
---
apiVersion: v1
kind: Service
metadata:
  name: example-grpc-server
  namespace: example-grpc-server
spec:
  selector:
    app: example-grpc-server
  ports:
  - protocol: TCP
    port: 50051
    targetPort: 50051
  type: ClusterIP
EOF
```

Example grpc server should be running on addr example-grpc-server:50051

#### 3. Deploy grpc-rest-proxy as standalone Kubernetes deployment

```bash
helm upgrade \
    --install \
    --namespace example-grpc-server \
    --set grpcRestProxy.grpcTargetAddr=example-grpc-server:50051 \
    --set image.repository=${GRPC_REST_PROXY_CONTAINER_REPOSITORY}/grpc-rest-proxy \
    --set image.tag=latest \
    grpc-rest-proxy ./charts/grpc-rest-proxy
```

#### 4. Send HTTP request to example-grpc-server through grpc-rest-proxy

```bash
POD_NAME=$(kubectl get pods --namespace example-grpc-server -l "app.kubernetes.io/name=grpc-rest-proxy,app.kubernetes.io/instance=grpc-rest-proxy" -o jsonpath="{.items[0].metadata.name}")
CONTAINER_PORT=$(kubectl get pod --namespace example-grpc-server $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
kubectl --namespace example-grpc-server port-forward $POD_NAME 8080:$CONTAINER_PORT

curl http://127.0.0.1:8080/api/users/John
```
