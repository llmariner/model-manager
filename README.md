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

Run the following command and run `loader`. Please note that it is better to run this on
an EC2 instance as it requires download and upload of large files.

```bash
python3 -m venv ./venv
source ./venv/bin/activate
pip install -U "huggingface_hub[cli]"

export AWS_PROFILE=<profile that has access to the bucket>
export HUGGING_FACE_HUB_TOKEN=<Hugging Face API key>

make build-loader

cat << EOF > config.yaml
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
    cacheDir: /home/ubuntu/.cache/huggingface/hub

debug:
  standalone: true
EOF

./bin/loader run --config config.yaml
```

### Generating a GGUF file

There might not be a GGUF file in Hugging Face repositories. If so, run the following
command to convert:

```baash
pip install numpy
pip install torch
pip install sentencepiece
pip install safetensors
pip install transformers

MODEL_NAME=meta-llama/Meta-Llama-3-8B-Instruct
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp

mkdir hf-model-dir
huggingface-cli download "${MODEL_NAME}" --local-dir=hf-model-dir
python3 convert_hf_to_gguf.py --outtype=f32 ./hf-model-dir --outfile model.gguf
mv model.gguf hf-model-dir/

aws s3 cp --recursive ./hf-model-dir s3://llm-operator-models/v1/base-models/"${MODEL_NAME}"
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
make llama-quantize

ORIG_MODEL_PATH=./hf-model-dir
python convert_hf_to_gguf.py ${ORIG_MODEL_PATH} --outtype f16 --outfile converted.bin
# See https://github.com/ggerganov/llama.cpp/discussions/406 to understand options like q4_0.
./llama-quantize converted.bin quantized.bin q4_0

MODEL_NAME=<target model name (e.g., meta-llama/Meta-Llama-3.1-8B-Instruct-q4 )>
aws s3 cp quantized.bin s3://llm-operator-models/v1/base-models/"${MODEL_NAME}"/model.gguf
```
