# iam-runtime-infratographer - An Infratographer IAM runtime implementation

iam-runtime-infratographer is an [IAM runtime][iam-runtime] implementation that uses [identity-api][identity-api] for authenticating subjects and [permissions-api][permissions-api] for checking access to resources. This allows applications to make use of Infratographer IAM functionality without needing to include dependencies directly in application code or mock services in development.

[iam-runtime]: https://github.com/metal-toolbox/iam-runtime
[identity-api]: https://github.com/infratographer/identity-api
[permissions-api]: https://github.com/infratographer/permissions-api

## Usage

iam-runtime-infratographer can be run as a standalone binary or a sidecar in a Kubernetes deployment.

To run it as a standalone binary using the provided example config, use the following commands:

```
$ make build # macOS users may need to run "GOOS=darwin make build"
$ ./iam-runtime-infratographer serve --config config.example.yaml
```

## Configuration

iam-runtime-infratographer can be configured using either a config file, command line arguments, or environment variables. An example config file is located at config.example.yaml.

## Example Kubernetes deployment

Below provides an example of adding the IAM runtime as a sidecar to your app deployment.

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: iam-runtime-config
data:
  config.yaml: |
    server:
      socketpath: /var/iam-runtime/runtime.sock
    permissions:
      host: permissions-api.internal.enterprise.net
    jwt:
      jwksuri: https://iam.example.com/jwks.json
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: app
            image: example/my-app:latest
            volumeMounts:
              - name: iam-runtime-socket
                mountPath: /var/iam-runtime/
        - name: iam-runtime
            image: ghcr.io/infratographer/iam-runtime-infratographer:v0.1.0
            volumeMounts:
              - name: iam-runtime-config
                mountPath: /etc/iam-runtime-infratographer/
              - name: iam-runtime-socket
                mountPath: /var/iam-runtime/
      volumes:
        - name: iam-runtime-config
          configMap:
            name: iam-runtime-config
        - name: iam-runtime-socket
            emptyDir: {}
```
