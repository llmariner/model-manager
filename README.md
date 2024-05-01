# Model Manager

Model Manager consists of the following components;

- `server`: Serve gRPC requests and HTTP requests
- `loadedr`: Load open models to the system

`loader` currently loads open models from Hugging Face, but we can extend that to support other locations.

## Running with Docker Compose

Run the following command:

```bash
docker-compose build
docker-compose up
```

You can access to the database or hit the HTTP endpoint:

```bash
docker exec -it <postgres container ID> psql -h localhost -U user --no-password -p 5432 -d model_manager

curl http://localhost:8080/v1/models

docker exec -it <aws-cli container ID> bash
export AWS_ACCESS_KEY_ID=llm-operator-key
export AWS_SECRET_ACCESS_KEY=llm-operator-secret
aws --endpoint-url http://minio:9000 s3 ls s3://llm-operator
```

## Running `server` Locally

```bash
make build-server
./bin/server run --config config.yaml
```

`config.yaml` has the following content:

```yaml
httpPort: 8080
grpcPort: 8081
internalGrpcPort: 8082

objectStore:
  s3:
    pathPrefix: models

debug:
  standalone: true
  sqlitePath: /tmp/model_manager.db
```

You can then connect to the DB.

```bash
sqlite3 /tmp/model_manager.db
# Run the query inside the database.
insert into models
  (model_id, tenant_id, created_at, updated_at)
values
  ('my-model', 'fake-tenant-id', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
```

You can then hit the endpoint.

```bash
curl http://localhost:8080/v1/models

grpcurl -d '{"base_model": "base", "suffix": "suffix", "tenant_id": "fake-tenant-id"}' -plaintext localhost:8082 list llmoperator.models.server.v1.ModelsInternalService/CreateModel
```

# Uploading models to S3 bucket `llm-operator-models`

Run `loader` with the following `config.yaml`:

```console
$ cat config.yaml

objectStore:
  s3:
    endpointUrl: https://s3.us-west-2.amazonaws.com
    region: us-west-2
    bucket: llm-operator-models
    pathPrefix: v1
    baseModelPathPrefix: base-models

baseModels:
- google/gemma-2b

runOnce: true

downloader:
  kind: huggingFace
  huggingFace:
    # Change this to your cache directory.
    cacheDir: /Users/kenji/.cache/huggingface/hub

debug:
  standalone: true

$ export AWS_PROFILE=<profile that has access to the bucket>
$ ./bin/loader run --config config.yaml
```

### Quantizing

See https://github.com/ggerganov/llama.cpp/discussions/2948 and
https://github.com/ollama/ollama/blob/main/docs/import.md.

```bash
make build-docker-convert-gguf

# Mount the volume where a original model is stored (without symlink).
docker run \
  -it \
  --entrypoint /bin/bash \
  -v /Users/kenji/base-models:/base-models \
  llm-operator/experiments-convert_gguf:latest

python convert.py /base-models --outfile google-gemma-2b-q8_0 --outtype q8_0
```

Here is another example:

```bash
git clone https://github.com/ggerganov/llama.cpp
cd llama.cp
make quantize

python convert-hf-to-gguf.py <model-path> --outtype f16 --outfile converted.bin
./quantize converted.bin quantized.bin q4_0
```
