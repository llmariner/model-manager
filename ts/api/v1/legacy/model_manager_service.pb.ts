/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../../fetch.pb"
import * as LlmarinerModelsServerV1Model_manager_service from "../model_manager_service.pb"
export class ModelsWorkerService {
  static CreateStorageConfig(req: LlmarinerModelsServerV1Model_manager_service.CreateStorageConfigRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.StorageConfig> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.CreateStorageConfigRequest, LlmarinerModelsServerV1Model_manager_service.StorageConfig>(`/llmoperator.models.server.v1.ModelsWorkerService/CreateStorageConfig`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetStorageConfig(req: LlmarinerModelsServerV1Model_manager_service.GetStorageConfigRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.StorageConfig> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.GetStorageConfigRequest, LlmarinerModelsServerV1Model_manager_service.StorageConfig>(`/llmoperator.models.server.v1.ModelsWorkerService/GetStorageConfig`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModel(req: LlmarinerModelsServerV1Model_manager_service.GetModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.Model> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.GetModelRequest, LlmarinerModelsServerV1Model_manager_service.Model>(`/llmoperator.models.server.v1.ModelsWorkerService/GetModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static RegisterModel(req: LlmarinerModelsServerV1Model_manager_service.RegisterModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.RegisterModelResponse> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.RegisterModelRequest, LlmarinerModelsServerV1Model_manager_service.RegisterModelResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/RegisterModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static PublishModel(req: LlmarinerModelsServerV1Model_manager_service.PublishModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.PublishModelResponse> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.PublishModelRequest, LlmarinerModelsServerV1Model_manager_service.PublishModelResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/PublishModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModelPath(req: LlmarinerModelsServerV1Model_manager_service.GetModelPathRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.GetModelPathResponse> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.GetModelPathRequest, LlmarinerModelsServerV1Model_manager_service.GetModelPathResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/GetModelPath`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetModelAttributes(req: LlmarinerModelsServerV1Model_manager_service.GetModelAttributesRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.ModelAttributes> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.GetModelAttributesRequest, LlmarinerModelsServerV1Model_manager_service.ModelAttributes>(`/llmoperator.models.server.v1.ModelsWorkerService/GetModelAttributes`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static CreateBaseModel(req: LlmarinerModelsServerV1Model_manager_service.CreateBaseModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.BaseModel> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.CreateBaseModelRequest, LlmarinerModelsServerV1Model_manager_service.BaseModel>(`/llmoperator.models.server.v1.ModelsWorkerService/CreateBaseModel`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
  static GetBaseModelPath(req: LlmarinerModelsServerV1Model_manager_service.GetBaseModelPathRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.GetBaseModelPathResponse> {
    return fm.fetchReq<LlmarinerModelsServerV1Model_manager_service.GetBaseModelPathRequest, LlmarinerModelsServerV1Model_manager_service.GetBaseModelPathResponse>(`/llmoperator.models.server.v1.ModelsWorkerService/GetBaseModelPath`, {...initReq, method: "POST", body: JSON.stringify(req)})
  }
}