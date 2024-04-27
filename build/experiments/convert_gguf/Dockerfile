FROM --platform=$BUILDPLATFORM python:3.12-slim
ARG TARGETARCH

WORKDIR /run

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY convert.py .
COPY gguf-py .
