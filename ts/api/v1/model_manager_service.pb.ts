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

export enum ActivationStatus {
  ACTIVATION_STATUS_UNSPECIFIED = "ACTIVATION_STATUS_UNSPECIFIED",
  ACTIVATION_STATUS_ACTIVE = "ACTIVATION_STATUS_ACTIVE",
  ACTIVATION_STATUS_INACTIVE = "ACTIVATION_STATUS_INACTIVE",
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

export type ModelConfigRuntimeConfigResources = {
  gpu?: number
}

export type ModelConfigRuntimeConfig = {
  resources?: ModelConfigRuntimeConfigResources
  replicas?: number
  extra_args?: string[]
}

export type ModelConfigClusterAllocationPolicy = {
  allowed_cluster_ids?: string[]
}

export type ModelConfig = {
  runtime_config?: ModelConfigRuntimeConfig
  cluster_allocation_policy?: ModelConfigClusterAllocationPolicy
}

export type Model = {
  id?: string
  created?: string
  object?: string
  owned_by?: string
  loading_status?: ModelLoadingStatus
  source_repository?: SourceRepository
  loading_failure_reason?: string
  formats?: ModelFormat[]
  is_base_model?: boolean
  base_model_id?: string
  activation_status?: ActivationStatus
  config?: ModelConfig
}

export type CreateModelRequest = {
  id?: string
  source_repository?: SourceRepository
  is_fine_tuned_model?: boolean
  base_model_id?: string
  suffix?: string
  model_file_location?: string
  config?: ModelConfig
  is_project_scoped?: boolean
}

export type ListModelsRequest = {
  include_loading_models?: boolean
  after?: string
  limit?: number
}

export type ListModelsResponse = {
  object?: string
  data?: Model[]
  has_more?: boolean
  total_items?: number
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

export type ActivateModelRequest = {
  id?: string
}

export type ActivateModelResponse = {
}

export type DeactivateModelRequest = {
  id?: string
}

export type DeactivateModelResponse = {
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
  project_id?: string
}

export type BaseModel = {
  id?: string
  created?: string
  object?: string
}

export type GetBaseModelPathRequest = {
  id?: string
  project_id?: string
}

export type GetBaseModelPathResponse = {
  formats?: ModelFormat[]
  path?: string
  gguf_model_path?: string
}

export type CreateHFModelRepoRequest = {
  name?: string
  project_id?: string
}

export type HFModelRepo = {
  name?: string
}

export type GetHFModelRepoRequest = {
  name?: string
  project_id?: string
}

export type AcquireUnloadedBaseModelRequest = {
}

export type AcquireUnloadedBaseModelResponse = {
  base_model_id?: string
  source_repository?: SourceRepository
  project_id?: string
}

export type UpdateBaseModelLoadingStatusRequestSuccess = {
}

export type UpdateBaseModelLoadingStatusRequestFailure = {
  reason?: string
}


type BaseUpdateBaseModelLoadingStatusRequest = {
  id?: string
  project_id?: string
}

export type UpdateBaseModelLoadingStatusRequest = BaseUpdateBaseModelLoadingStatusRequest
  & OneOf<{ success: UpdateBaseModelLoadingStatusRequestSuccess; failure: UpdateBaseModelLoadingStatusRequestFailure }>

export type UpdateBaseModelLoadingStatusResponse = {
}

export type AcquireUnloadedModelRequest = {
}

export type AcquireUnloadedModelResponse = {
  model_id?: string
  is_base_model?: boolean
  source_repository?: SourceRepository
  model_file_location?: string
  dest_path?: string
}

export type UpdateModelLoadingStatusRequestSuccess = {
}

export type UpdateModelLoadingStatusRequestFailure = {
  reason?: string
}


type BaseUpdateModelLoadingStatusRequest = {
  id?: string
}

export type UpdateModelLoadingStatusRequest = BaseUpdateModelLoadingStatusRequest
  & OneOf<{ success: UpdateModelLoadingStatusRequestSuccess; failure: UpdateModelLoadingStatusRequestFailure }>

export type UpdateModelLoadingStatusResponse = {
}

export class ModelsService {
  static ListModels(req: ListModelsRequest, initReq?: fm.InitReq): Promise<ListModelsResponse> {
    return fm.fetchReq<ListModelsRequest, ListModelsResponse>(`/v1/models?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
  static GetModel(req: GetModelRequest, initReq?: fm.InitReq): Promise<Model> {
    return fm.fetchReq<GetModelRequest, Model>(`/v1/models/${req["id==**"]}?${fm.renderURLSearchParams(req, ["id==**"])}`, {...initReq, method: "GET"})
  }
  static DeleteModel(req: DeleteModelRequest, initReq?: fm.InitReq): Promise<DeleteModelResponse> {
    return fm.fetchReq<DeleteModelRequest, DeleteModelResponse>(`/v1/models/${req["id=**"]}`, {...initReq, method: "DELETE"})
  }
  static CreateModel(req: CreateModelRequest, initReq?: fm.InitReq): Promise<Model> {
    return fm.fetchReq<CreateModelRequest, Model>(`/v1/models`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static ActivateModel(req: ActivateModelRequest, initReq?: fm.InitReq): Promise<ActivateModelResponse> {
    return fm.fetchReq<ActivateModelRequest, ActivateModelResponse>(`/v1/models/${req["id"]}:activate`, {...initReq, method: "POST"})
  }
  static DeactivateModel(req: DeactivateModelRequest, initReq?: fm.InitReq): Promise<DeactivateModelResponse> {
    return fm.fetchReq<DeactivateModelRequest, DeactivateModelResponse>(`/v1/models/${req["id"]}:deactivate`, {...initReq, method: "POST"})
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
  static ListModels(req: ListModelsRequest, initReq?: fm.InitReq): Promise<ListModelsResponse> {
    return fm.fetchReq<ListModelsRequest, ListModelsResponse>(`/llmariner.models.server.v1.ModelsWorkerService/ListModels`, {...initReq, method: "POST", body: JSON.stringify(req)})
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
  static AcquireUnloadedModel(req: AcquireUnloadedModelRequest, initReq?: fm.InitReq): Promise<AcquireUnloadedModelResponse> {
    return fm.fetchReq<AcquireUnloadedModelRequest, AcquireUnloadedModelResponse>(`/llmariner.models.server.v1.ModelsWorkerService/AcquireUnloadedModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static UpdateBaseModelLoadingStatus(req: UpdateBaseModelLoadingStatusRequest, initReq?: fm.InitReq): Promise<UpdateBaseModelLoadingStatusResponse> {
    return fm.fetchReq<UpdateBaseModelLoadingStatusRequest, UpdateBaseModelLoadingStatusResponse>(`/llmariner.models.server.v1.ModelsWorkerService/UpdateBaseModelLoadingStatus`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static UpdateModelLoadingStatus(req: UpdateModelLoadingStatusRequest, initReq?: fm.InitReq): Promise<UpdateModelLoadingStatusResponse> {
    return fm.fetchReq<UpdateModelLoadingStatusRequest, UpdateModelLoadingStatusResponse>(`/llmariner.models.server.v1.ModelsWorkerService/UpdateModelLoadingStatus`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}