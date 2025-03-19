/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);

export enum ModelFormat {
  MODEL_FORMAT_UNSPECIFIED = "MODEL_FORMAT_UNSPECIFIED",
  MODEL_FORMAT_GGUF = "MODEL_FORMAT_GGUF",
  MODEL_FORMAT_HUGGING_FACE = "MODEL_FORMAT_HUGGING_FACE",
  MODEL_FORMAT_NVIDIA_TRITON = "MODEL_FORMAT_NVIDIA_TRITON",
  MODEL_FORMAT_OLLAMA = "MODEL_FORMAT_OLLAMA",
}

export enum ModelLoadingStatus {
  MODEL_LOADING_STATUS_UNSPECIFIED = "MODEL_LOADING_STATUS_UNSPECIFIED",
  MODEL_LOADING_STATUS_REQUESTED = "MODEL_LOADING_STATUS_REQUESTED",
  MODEL_LOADING_STATUS_LOADING = "MODEL_LOADING_STATUS_LOADING",
  MODEL_LOADING_STATUS_SUCCEEDED = "MODEL_LOADING_STATUS_SUCCEEDED",
  MODEL_LOADING_STATUS_FAILED = "MODEL_LOADING_STATUS_FAILED",
}

export enum SourceRepository {
  SOURCE_REPOSITORY_UNSPECIFIED = "SOURCE_REPOSITORY_UNSPECIFIED",
  SOURCE_REPOSITORY_OBJECT_STORE = "SOURCE_REPOSITORY_OBJECT_STORE",
  SOURCE_REPOSITORY_HUGGING_FACE = "SOURCE_REPOSITORY_HUGGING_FACE",
  SOURCE_REPOSITORY_OLLAMA = "SOURCE_REPOSITORY_OLLAMA",
  SOURCE_REPOSITORY_FINE_TUNING = "SOURCE_REPOSITORY_FINE_TUNING",
}

export enum AdapterType {
  ADAPTER_TYPE_UNSPECIFIED = "ADAPTER_TYPE_UNSPECIFIED",
  ADAPTER_TYPE_LORA = "ADAPTER_TYPE_LORA",
  ADAPTER_TYPE_QLORA = "ADAPTER_TYPE_QLORA",
}

export enum QuantizationType {
  QUANTIZATION_TYPE_UNSPECIFIED = "QUANTIZATION_TYPE_UNSPECIFIED",
  QUANTIZATION_TYPE_GGUF = "QUANTIZATION_TYPE_GGUF",
  QUANTIZATION_TYPE_AWQ = "QUANTIZATION_TYPE_AWQ",
}

export type ModelFormats = {
  formats?: ModelFormat[]
}

export type Model = {
  id?: string
  created?: string
  object?: string
  owned_by?: string
  loading_status?: ModelLoadingStatus
  source_repository?: SourceRepository
  loading_failure_reason?: string
}

export type CreateModelRequest = {
  id?: string
  source_repository?: SourceRepository
}

export type ListModelsRequest = {
  include_loading_models?: boolean
}

export type ListModelsResponse = {
  object?: string
  data?: Model[]
}

export type GetModelRequest = {
  id?: string
  include_loading_model?: boolean
}

export type DeleteModelRequest = {
  id?: string
}

export type DeleteModelResponse = {
  id?: string
  object?: string
  deleted?: boolean
}

export type ListBaseModelsRequest = {
}

export type BaseModel = {
  id?: string
  created?: string
  object?: string
}

export type ListBaseModelsResponse = {
  object?: string
  data?: BaseModel[]
}

export type StorageConfig = {
  path_prefix?: string
}

export type CreateStorageConfigRequest = {
  path_prefix?: string
}

export type GetStorageConfigRequest = {
}

export type RegisterModelRequest = {
  id?: string
  base_model?: string
  suffix?: string
  organization_id?: string
  project_id?: string
  adapter?: AdapterType
  quantization?: QuantizationType
  path?: string
}

export type RegisterModelResponse = {
  id?: string
  path?: string
}

export type PublishModelRequest = {
  id?: string
}

export type PublishModelResponse = {
}

export type GetModelPathRequest = {
  id?: string
}

export type GetModelPathResponse = {
  path?: string
}

export type ModelAttributes = {
  path?: string
  base_model?: string
  adapter?: AdapterType
  quantization?: QuantizationType
}

export type GetModelAttributesRequest = {
  id?: string
}

export type CreateBaseModelRequest = {
  id?: string
  path?: string
  formats?: ModelFormat[]
  gguf_model_path?: string
  source_repository?: SourceRepository
}

export type GetBaseModelPathRequest = {
  id?: string
}

export type GetBaseModelPathResponse = {
  formats?: ModelFormat[]
  path?: string
  gguf_model_path?: string
}

export type CreateHFModelRepoRequest = {
  name?: string
}

export type HFModelRepo = {
  name?: string
}

export type GetHFModelRepoRequest = {
  name?: string
}

export type AcquireUnloadedBaseModelRequest = {
}

export type AcquireUnloadedBaseModelResponse = {
  base_model_id?: string
  source_repository?: SourceRepository
}

export type UpdateBaseModelLoadingStatusRequestSuccess = {
}

export type UpdateBaseModelLoadingStatusRequestFailure = {
  reason?: string
}


type BaseUpdateBaseModelLoadingStatusRequest = {
  id?: string
}

export type UpdateBaseModelLoadingStatusRequest = BaseUpdateBaseModelLoadingStatusRequest
  & OneOf<{ success: UpdateBaseModelLoadingStatusRequestSuccess; failure: UpdateBaseModelLoadingStatusRequestFailure }>

export type UpdateBaseModelLoadingStatusResponse = {
}

export class ModelsService {
  static ListModels(req: ListModelsRequest, initReq?: fm.InitReq): Promise<ListModelsResponse> {
    return fm.fetchReq<ListModelsRequest, ListModelsResponse>(`/v1/models?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetModel(req: GetModelRequest, initReq?: fm.InitReq): Promise<Model> {
    return fm.fetchReq<GetModelRequest, Model>(`/v1/models/${req["id"]}?${fm.renderURLSearchParams(req, ["id"])}`, {...initReq, method: "GET"})
  }
  static DeleteModel(req: DeleteModelRequest, initReq?: fm.InitReq): Promise<DeleteModelResponse> {
    return fm.fetchReq<DeleteModelRequest, DeleteModelResponse>(`/v1/models/${req["id"]}`, {...initReq, method: "DELETE"})
  }
  static CreateModel(req: CreateModelRequest, initReq?: fm.InitReq): Promise<Model> {
    return fm.fetchReq<CreateModelRequest, Model>(`/v1/models`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ListBaseModels(req: ListBaseModelsRequest, initReq?: fm.InitReq): Promise<ListBaseModelsResponse> {
    return fm.fetchReq<ListBaseModelsRequest, ListBaseModelsResponse>(`/v1/basemodels?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}
export class ModelsWorkerService {
  static CreateStorageConfig(req: CreateStorageConfigRequest, initReq?: fm.InitReq): Promise<StorageConfig> {
    return fm.fetchReq<CreateStorageConfigRequest, StorageConfig>(`/llmariner.models.server.v1.ModelsWorkerService/CreateStorageConfig`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetStorageConfig(req: GetStorageConfigRequest, initReq?: fm.InitReq): Promise<StorageConfig> {
    return fm.fetchReq<GetStorageConfigRequest, StorageConfig>(`/llmariner.models.server.v1.ModelsWorkerService/GetStorageConfig`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModel(req: GetModelRequest, initReq?: fm.InitReq): Promise<Model> {
    return fm.fetchReq<GetModelRequest, Model>(`/llmariner.models.server.v1.ModelsWorkerService/GetModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static RegisterModel(req: RegisterModelRequest, initReq?: fm.InitReq): Promise<RegisterModelResponse> {
    return fm.fetchReq<RegisterModelRequest, RegisterModelResponse>(`/llmariner.models.server.v1.ModelsWorkerService/RegisterModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static PublishModel(req: PublishModelRequest, initReq?: fm.InitReq): Promise<PublishModelResponse> {
    return fm.fetchReq<PublishModelRequest, PublishModelResponse>(`/llmariner.models.server.v1.ModelsWorkerService/PublishModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModelPath(req: GetModelPathRequest, initReq?: fm.InitReq): Promise<GetModelPathResponse> {
    return fm.fetchReq<GetModelPathRequest, GetModelPathResponse>(`/llmariner.models.server.v1.ModelsWorkerService/GetModelPath`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModelAttributes(req: GetModelAttributesRequest, initReq?: fm.InitReq): Promise<ModelAttributes> {
    return fm.fetchReq<GetModelAttributesRequest, ModelAttributes>(`/llmariner.models.server.v1.ModelsWorkerService/GetModelAttributes`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateBaseModel(req: CreateBaseModelRequest, initReq?: fm.InitReq): Promise<BaseModel> {
    return fm.fetchReq<CreateBaseModelRequest, BaseModel>(`/llmariner.models.server.v1.ModelsWorkerService/CreateBaseModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetBaseModelPath(req: GetBaseModelPathRequest, initReq?: fm.InitReq): Promise<GetBaseModelPathResponse> {
    return fm.fetchReq<GetBaseModelPathRequest, GetBaseModelPathResponse>(`/llmariner.models.server.v1.ModelsWorkerService/GetBaseModelPath`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateHFModelRepo(req: CreateHFModelRepoRequest, initReq?: fm.InitReq): Promise<HFModelRepo> {
    return fm.fetchReq<CreateHFModelRepoRequest, HFModelRepo>(`/llmariner.models.server.v1.ModelsWorkerService/CreateHFModelRepo`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetHFModelRepo(req: GetHFModelRepoRequest, initReq?: fm.InitReq): Promise<HFModelRepo> {
    return fm.fetchReq<GetHFModelRepoRequest, HFModelRepo>(`/llmariner.models.server.v1.ModelsWorkerService/GetHFModelRepo`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static AcquireUnloadedBaseModel(req: AcquireUnloadedBaseModelRequest, initReq?: fm.InitReq): Promise<AcquireUnloadedBaseModelResponse> {
    return fm.fetchReq<AcquireUnloadedBaseModelRequest, AcquireUnloadedBaseModelResponse>(`/llmariner.models.server.v1.ModelsWorkerService/AcquireUnloadedBaseModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static UpdateBaseModelLoadingStatus(req: UpdateBaseModelLoadingStatusRequest, initReq?: fm.InitReq): Promise<UpdateBaseModelLoadingStatusResponse> {
    return fm.fetchReq<UpdateBaseModelLoadingStatusRequest, UpdateBaseModelLoadingStatusResponse>(`/llmariner.models.server.v1.ModelsWorkerService/UpdateBaseModelLoadingStatus`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}