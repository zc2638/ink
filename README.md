# ink

Controllable CICD service

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
