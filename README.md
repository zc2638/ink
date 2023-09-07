# ink
![LICENSE](https://img.shields.io/github/license/zc2638/swag.svg?style=flat-square&color=blue)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/zc2638/ink/main.yaml?branch=main&style=flat-square)](https://github.com/zc2638/ink/actions/workflows/main.yaml)

Controllable CICD service

## TODO

- Add the worker for kubernetes.
- Support docker image pull auth for registry.
- Add more storage backend support, like MySQL and Postgres.
- Support setting mode.

## Setup

### 一、Source

#### 1. Run inkd

```shell
go run ./cmd/inkd --config config/config.yaml
```

#### 2. Run inker

```shell
go run ./cmd/inker --config config/config.yaml --config-sub-key worker
```

#### 3. Install inkctl

```shell
go install github.com/zc2638/ink/cmd/inkctl@latest
```

bash completion

```shell
source <(inkctl completion bash)
```

## Resources

### Workflow

For detailed structure, please go to: [v1.Workflow](./pkg/api/core/v1/workflow.go)

#### Example

```yaml
kind: Workflow
name: test-docker
namespace: default
spec:
  steps:
    - name: step1
      image: alpine:3.18
      command:
        - echo 'step1 docker' > test.log
    - name: step2
      image: alpine:3.18
      shell:
        - /bin/bash
        - -c
      command:
        - echo "step2 sleep"
        - sleep 5
        - echo "step2 sleep over"
    - name: step3
      image: alpine:3.18
      env:
        - name: STATUS
          value: success
      command:
        - cat test.log
        - echo "step3 $STATUS"
        - pwd
```

### Box

For detailed structure, please go to: [v1.Box](./pkg/api/core/v1/box.go)

#### Example

```yaml
kind: Box
name: test
namespace: default
resources:
  - name: test-docker
    kind: Workflow
  - name: test-secret
    kind: Secret
```

### Secret

For detailed structure, please go to: [v1.Secret](./pkg/api/core/v1/secret.go)

#### Example

```yaml
kind: Secret
name: test-secret
namespace: default
data:
  secret1: secret1abc123
  secret2: this is secret2
```
