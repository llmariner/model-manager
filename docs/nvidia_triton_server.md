# Note on Nvidia Triton Server

This is a short note on how to compile a model that Nvidia Triton Server can process.
This mostly follows https://www.infracloud.io/blogs/running-llama-3-with-triton-tensorrt-llm/

## Software/library versions

We use the following versions:

- [TensorRT-LLM](https://github.com/NVIDIA/TensorRT-LLM.git): v0.13.0
- [Triton backend for TensorRT-LLM](https://github.com/triton-inference-server/tensorrtllm_backend.git): v0.13.0
- [Triton Inference Server](https://github.com/triton-inference-server/server): 2.50.0 (corresponding to NGC container 24.09)

Please note that you need to use a specific version of TensorRT-LLM

[The release note of Triton Inference Server](https://github.com/triton-inference-server/server/releases/tag/v2.50.0) says Tensor-RT LLM v0.13.0 is compatible with 24.07 (not 24.09), but that looks like a documentation again.

As written in the above note, the Python transformer libraries need to be 4.43+ to run Llama 3.1.

## Model compilation

To serve a model, we need to convert model files available from Hugging Face to a specific format that Triton Inference Server can process.

In this example, we take the following three steps:

1. Quantize the Llama3.1 8B instruct model with AWQ
2. Convert the model to a format that TensortRT can accept.
3. Build a model repository for Triton Inference Server by creating additional configuration files


### Step 1. Quantize the Llama3.1 8B instruct model with AWQ

First download the model from Hugging Face.

```console
mkdir ./triton
cd triton

export HUGGING_FACE_HUB_TOKEN=<your token>
huggingface-cli download meta-llama/Meta-Llama-3.1-8B-Instruct \
  --local-dir=Meta-Llama-3-8B-Instruct/ \
  --local-dir-use-symlink=False
```

Also clone the TensortRT-LLM repository.

```console
git clone -b v0.13.0 https://github.com/NVIDIA/TensorRT-LLM.git
```

Then start a Ubuntu container that has the CUDA toolkit.

```console
docker run --rm --gpus all \
  --entrypoint /bin/bash \
  --ipc=host \
  --ulimit memlock=-1 \
  --shm-size=20g \
  --volume $(pwd):/data \
  --workdir /data \
  -it nvidia/cuda:12.4.0-devel-ubuntu22.04
```

Run the following inside the container to quantize the model.

```console
apt-get update && apt-get -y install python3.10 python3-pip openmpi-bin libopenmpi-dev
pip3 install tensorrt_llm==0.13.0 -U --extra-index-url https://pypi.nvidia.com
pip install  nvidia-ammo~=0.7.3 --no-cache-dir --extra-index-url https://pypi.nvidia.com

cd /data/TensorRT-LLM/examples/quantization

# Note: Install specific version of Cython and pyaml first to work around https://stackoverflow.com/questions/77490435/attributeerror-cython-sources
pip install "Cython<3.0" "pyyaml<6" --no-build-isolation
pip install -r requirements.txt

# Upgrade transfomers to fix https://huggingface.co/meta-llama/Llama-3.1-8B-Instruct/discussions/15
pip install --upgrade transformers

python3 quantize.py \
  --model_dir /data/Meta-Llama-3-8B-Instruct \
  --output_dir /data/Meta-Llama-3-8B-Instruct-awq \
  --dtype bfloat16 \
  --qformat int4_awq \
  --awq_block_size 128 \
  --batch_size 12
```


### Step 2. Convert the model to a format that TensortRT can accept.

Run the following command inside the container to convert the model.

```console
trtllm-build \
  --checkpoint_dir /data/Meta-Llama-3-8B-Instruct-awq \
  --gpt_attention_plugin bfloat16 \
  --gemm_plugin bfloat16 \
  --output_dir /data/llama3-engine
```

You can test the model by running the following Python script.

```console
python3 ../run.py \
  --engine_dir=/data/llama3-engine \
  --max_output_len 200 \
  --tokenizer_dir /data/Meta-Llama-3-8B-Instruct \
  --input_text "How do I count to nine in Hindi?"
```

After this, you can stop the Ubuntu container.

### Step 3. Build a model repository for Triton Inference Server

Create a model registry directory and copy necessary files from the TensorRT-LLM backend repository.

```console
mkdir -p repo/llama3

git clone -b v0.13.0 https://github.com/triton-inference-server/tensorrtllm_backend.git
cp -r tensorrtllm_backend/all_models/inflight_batcher_llm/* repo/llama3/
```

Then copy the files generated in the previous to the model registry directory and
remove the unnecessary `tensorrt_llm_bls` directory.

```consolde
cp llama3-engine/* repo/llama3/tensorrt_llm/1/
rm -r repo/llama3/tensorrt_llm_bls
```

Create additional configuration files.

```console
cd /data/
# Set these to a path where the model file is available inside the triton server container.
HF_LLAMA_MODEL="/data/Meta-Llama-3-8B-Instruct"
ENGINE_PATH="/data/repo/llama3/tensorrt_llm/1"

python3 tensorrtllm_backend/tools/fill_template.py -i repo/llama3/preprocessing/config.pbtxt tokenizer_dir:${HF_LLAMA_MODEL},tokenizer_type:auto,triton_max_batch_size:1,preprocessing_instance_count:1
python3 tensorrtllm_backend/tools/fill_template.py -i repo/llama3/postprocessing/config.pbtxt tokenizer_dir:${HF_LLAMA_MODEL},tokenizer_type:auto,triton_max_batch_size:1,postprocessing_instance_count:1
python3 tensorrtllm_backend/tools/fill_template.py -i repo/llama3/ensemble/config.pbtxt triton_max_batch_size:1
python3 tensorrtllm_backend/tools/fill_template.py -i repo/llama3/tensorrt_llm/config.pbtxt triton_backend:tensorrtllm,triton_max_batch_size:1,decoupled_mode:False,max_beam_width:1,engine_dir:${ENGINE_PATH},max_tokens_in_paged_kv_cache:2560,max_attention_window_size:2560,kv_cache_free_gpu_mem_fraction:0.5,exclude_input_in_output:True,enable_kv_cache_reuse:False,batching_strategy:inflight_fused_batching,max_queue_delay_microseconds:0
```


## Serve inference requests

You can test the model registry by starting Triton Inference Server.

```console
docker run --rm -it --net host --gpus all \
  --shm-size=2g --ulimit memlock=-1 --ulimit stack=67108864 \
  -v $(pwd):/data \
  nvcr.io/nvidia/tritonserver:24.09-trtllm-python-py3 \
  tritonserver --model-repository=/data/repo/llama3
```

Here is an example curl command to hit the endpoint.

```console
curl -X POST http://localhost:8000/v2/models/ensemble/generate -d \
  '{"text_input": "What is machine learning?", "max_tokens": 20, "bad_words": "", "stop_words": ""}'
```
