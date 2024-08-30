/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"

export enum ModelFormat {
  MODEL_FORMAT_UNSPECIFIED = "MODEL_FORMAT_UNSPECIFIED",
  MODEL_FORMAT_GGUF = "MODEL_FORMAT_GGUF",
  MODEL_FORMAT_HUGGING_FACE = "MODEL_FORMAT_HUGGING_FACE",
}

export type Model = {
  id?: string
  created?: string
  object?: string
  ownedBy?: string
}

export type ListModelsRequest = {
}

export type ListModelsResponse = {
  object?: string
  data?: Model[]
}

export type GetModelRequest = {
  id?: string
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
  pathPrefix?: string
}

export type CreateStorageConfigRequest = {
  pathPrefix?: string
}

export type GetStorageConfigRequest = {
}

export type RegisterModelRequest = {
  baseModel?: string
  suffix?: string
  organizationId?: string
  projectId?: string
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

export type CreateBaseModelRequest = {
  id?: string
  path?: string
  format?: ModelFormat
  ggufModelPath?: string
}

export type GetBaseModelPathRequest = {
  id?: string
}

export type GetBaseModelPathResponse = {
  format?: ModelFormat
  path?: string
  ggufModelPath?: string
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
  static ListBaseModels(req: ListBaseModelsRequest, initReq?: fm.InitReq): Promise<ListBaseModelsResponse> {
    return fm.fetchReq<ListBaseModelsRequest, ListBaseModelsResponse>(`/v1/basemodels?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"})
  }
}
export class ModelsWorkerService {
  static CreateStorageConfig(req: CreateStorageConfigRequest, initReq?: fm.InitReq): Promise<StorageConfig> {
    return fm.fetchReq<CreateStorageConfigRequest, StorageConfig>(`/llmoperator.models.server.v1.ModelsWorkerService/CreateStorageConfig`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetStorageConfig(req: GetStorageConfigRequest, initReq?: fm.InitReq): Promise<StorageConfig> {
    return fm.fetchReq<GetStorageConfigRequest, StorageConfig>(`/llmoperator.models.server.v1.ModelsWorkerService/GetStorageConfig`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModel(req: GetModelRequest, initReq?: fm.InitReq): Promise<Model> {
    return fm.fetchReq<GetModelRequest, Model>(`/llmoperator.models.server.v1.ModelsWorkerService/GetModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static RegisterModel(req: RegisterModelRequest, initReq?: fm.InitReq): Promise<RegisterModelResponse> {
    return fm.fetchReq<RegisterModelRequest, RegisterModelResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/RegisterModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static PublishModel(req: PublishModelRequest, initReq?: fm.InitReq): Promise<PublishModelResponse> {
    return fm.fetchReq<PublishModelRequest, PublishModelResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/PublishModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModelPath(req: GetModelPathRequest, initReq?: fm.InitReq): Promise<GetModelPathResponse> {
    return fm.fetchReq<GetModelPathRequest, GetModelPathResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/GetModelPath`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateBaseModel(req: CreateBaseModelRequest, initReq?: fm.InitReq): Promise<BaseModel> {
    return fm.fetchReq<CreateBaseModelRequest, BaseModel>(`/llmoperator.models.server.v1.ModelsWorkerService/CreateBaseModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetBaseModelPath(req: GetBaseModelPathRequest, initReq?: fm.InitReq): Promise<GetBaseModelPathResponse> {
    return fm.fetchReq<GetBaseModelPathRequest, GetBaseModelPathResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/GetBaseModelPath`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}