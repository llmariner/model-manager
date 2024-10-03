import * as fm from "../../../fetch.pb";
import * as LlmarinerModelsServerV1Model_manager_service from "../model_manager_service.pb";
export declare class ModelsWorkerService {
    static CreateStorageConfig(req: LlmarinerModelsServerV1Model_manager_service.CreateStorageConfigRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.StorageConfig>;
    static GetStorageConfig(req: LlmarinerModelsServerV1Model_manager_service.GetStorageConfigRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.StorageConfig>;
    static GetModel(req: LlmarinerModelsServerV1Model_manager_service.GetModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.Model>;
    static RegisterModel(req: LlmarinerModelsServerV1Model_manager_service.RegisterModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.RegisterModelResponse>;
    static PublishModel(req: LlmarinerModelsServerV1Model_manager_service.PublishModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.PublishModelResponse>;
    static GetModelPath(req: LlmarinerModelsServerV1Model_manager_service.GetModelPathRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.GetModelPathResponse>;
    static GetModelAttributes(req: LlmarinerModelsServerV1Model_manager_service.GetModelAttributesRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.ModelAttributes>;
    static CreateBaseModel(req: LlmarinerModelsServerV1Model_manager_service.CreateBaseModelRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.BaseModel>;
    static GetBaseModelPath(req: LlmarinerModelsServerV1Model_manager_service.GetBaseModelPathRequest, initReq?: fm.InitReq): Promise<LlmarinerModelsServerV1Model_manager_service.GetBaseModelPathResponse>;
}
