import * as fm from "../../fetch.pb";
export declare enum ModelFormat {
    MODEL_FORMAT_UNSPECIFIED = "MODEL_FORMAT_UNSPECIFIED",
    MODEL_FORMAT_GGUF = "MODEL_FORMAT_GGUF",
    MODEL_FORMAT_HUGGING_FACE = "MODEL_FORMAT_HUGGING_FACE",
    MODEL_FORMAT_NVIDIA_TRITON = "MODEL_FORMAT_NVIDIA_TRITON"
}
export declare enum AdapterType {
    ADAPTER_TYPE_UNSPECIFIED = "ADAPTER_TYPE_UNSPECIFIED",
    ADAPTER_TYPE_LORA = "ADAPTER_TYPE_LORA",
    ADAPTER_TYPE_QLORA = "ADAPTER_TYPE_QLORA"
}
export declare enum QuantizationType {
    QUANTIZATION_TYPE_UNSPECIFIED = "QUANTIZATION_TYPE_UNSPECIFIED",
    QUANTIZATION_TYPE_GGUF = "QUANTIZATION_TYPE_GGUF",
    QUANTIZATION_TYPE_AWQ = "QUANTIZATION_TYPE_AWQ"
}
export type ModelFormats = {
    formats?: ModelFormat[];
};
export type Model = {
    id?: string;
    created?: string;
    object?: string;
    ownedBy?: string;
};
export type ListModelsRequest = {};
export type ListModelsResponse = {
    object?: string;
    data?: Model[];
};
export type GetModelRequest = {
    id?: string;
};
export type DeleteModelRequest = {
    id?: string;
};
export type DeleteModelResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type ListBaseModelsRequest = {};
export type BaseModel = {
    id?: string;
    created?: string;
    object?: string;
};
export type ListBaseModelsResponse = {
    object?: string;
    data?: BaseModel[];
};
export type StorageConfig = {
    pathPrefix?: string;
};
export type CreateStorageConfigRequest = {
    pathPrefix?: string;
};
export type GetStorageConfigRequest = {};
export type RegisterModelRequest = {
    id?: string;
    baseModel?: string;
    suffix?: string;
    organizationId?: string;
    projectId?: string;
    adapter?: AdapterType;
    quantization?: QuantizationType;
    path?: string;
};
export type RegisterModelResponse = {
    id?: string;
    path?: string;
};
export type PublishModelRequest = {
    id?: string;
};
export type PublishModelResponse = {};
export type GetModelPathRequest = {
    id?: string;
};
export type GetModelPathResponse = {
    path?: string;
};
export type ModelAttributes = {
    path?: string;
    baseModel?: string;
    adapter?: AdapterType;
    quantization?: QuantizationType;
};
export type GetModelAttributesRequest = {
    id?: string;
};
export type CreateBaseModelRequest = {
    id?: string;
    path?: string;
    formats?: ModelFormat[];
    ggufModelPath?: string;
};
export type GetBaseModelPathRequest = {
    id?: string;
};
export type GetBaseModelPathResponse = {
    formats?: ModelFormat[];
    path?: string;
    ggufModelPath?: string;
};
export declare class ModelsService {
    static ListModels(req: ListModelsRequest, initReq?: fm.InitReq): Promise<ListModelsResponse>;
    static GetModel(req: GetModelRequest, initReq?: fm.InitReq): Promise<Model>;
    static DeleteModel(req: DeleteModelRequest, initReq?: fm.InitReq): Promise<DeleteModelResponse>;
    static ListBaseModels(req: ListBaseModelsRequest, initReq?: fm.InitReq): Promise<ListBaseModelsResponse>;
}
export declare class ModelsWorkerService {
    static CreateStorageConfig(req: CreateStorageConfigRequest, initReq?: fm.InitReq): Promise<StorageConfig>;
    static GetStorageConfig(req: GetStorageConfigRequest, initReq?: fm.InitReq): Promise<StorageConfig>;
    static GetModel(req: GetModelRequest, initReq?: fm.InitReq): Promise<Model>;
    static RegisterModel(req: RegisterModelRequest, initReq?: fm.InitReq): Promise<RegisterModelResponse>;
    static PublishModel(req: PublishModelRequest, initReq?: fm.InitReq): Promise<PublishModelResponse>;
    static GetModelPath(req: GetModelPathRequest, initReq?: fm.InitReq): Promise<GetModelPathResponse>;
    static GetModelAttributes(req: GetModelAttributesRequest, initReq?: fm.InitReq): Promise<ModelAttributes>;
    static CreateBaseModel(req: CreateBaseModelRequest, initReq?: fm.InitReq): Promise<BaseModel>;
    static GetBaseModelPath(req: GetBaseModelPathRequest, initReq?: fm.InitReq): Promise<GetBaseModelPathResponse>;
}
