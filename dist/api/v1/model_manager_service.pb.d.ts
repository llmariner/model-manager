import * as fm from "../../fetch.pb";
type Absent<T, K extends keyof T> = {
    [k in Exclude<keyof T, K>]?: undefined;
};
type OneOf<T> = {
    [k in keyof T]?: undefined;
} | (keyof T extends infer K ? (K extends string & keyof T ? {
    [k in K]: T[K];
} & Absent<T, K> : never) : never);
export declare enum ModelFormat {
    MODEL_FORMAT_UNSPECIFIED = "MODEL_FORMAT_UNSPECIFIED",
    MODEL_FORMAT_GGUF = "MODEL_FORMAT_GGUF",
    MODEL_FORMAT_HUGGING_FACE = "MODEL_FORMAT_HUGGING_FACE",
    MODEL_FORMAT_NVIDIA_TRITON = "MODEL_FORMAT_NVIDIA_TRITON",
    MODEL_FORMAT_OLLAMA = "MODEL_FORMAT_OLLAMA"
}
export declare enum ModelLoadingStatus {
    MODEL_LOADING_STATUS_UNSPECIFIED = "MODEL_LOADING_STATUS_UNSPECIFIED",
    MODEL_LOADING_STATUS_REQUESTED = "MODEL_LOADING_STATUS_REQUESTED",
    MODEL_LOADING_STATUS_LOADING = "MODEL_LOADING_STATUS_LOADING",
    MODEL_LOADING_STATUS_SUCCEEDED = "MODEL_LOADING_STATUS_SUCCEEDED",
    MODEL_LOADING_STATUS_FAILED = "MODEL_LOADING_STATUS_FAILED"
}
export declare enum SourceRepository {
    SOURCE_REPOSITORY_UNSPECIFIED = "SOURCE_REPOSITORY_UNSPECIFIED",
    SOURCE_REPOSITORY_OBJECT_STORE = "SOURCE_REPOSITORY_OBJECT_STORE",
    SOURCE_REPOSITORY_HUGGING_FACE = "SOURCE_REPOSITORY_HUGGING_FACE",
    SOURCE_REPOSITORY_OLLAMA = "SOURCE_REPOSITORY_OLLAMA",
    SOURCE_REPOSITORY_FINE_TUNING = "SOURCE_REPOSITORY_FINE_TUNING"
}
export declare enum ActivationStatus {
    ACTIVATION_STATUS_UNSPECIFIED = "ACTIVATION_STATUS_UNSPECIFIED",
    ACTIVATION_STATUS_ACTIVE = "ACTIVATION_STATUS_ACTIVE",
    ACTIVATION_STATUS_INACTIVE = "ACTIVATION_STATUS_INACTIVE"
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
    owned_by?: string;
    loading_status?: ModelLoadingStatus;
    source_repository?: SourceRepository;
    loading_failure_reason?: string;
    formats?: ModelFormat[];
    is_base_model?: boolean;
    base_model_id?: string;
    activation_status?: ActivationStatus;
};
export type CreateModelRequest = {
    id?: string;
    source_repository?: SourceRepository;
    is_fine_tuned_model?: boolean;
    base_model_id?: string;
    suffix?: string;
    model_file_location?: string;
};
export type ListModelsRequest = {
    include_loading_models?: boolean;
    after?: string;
    limit?: number;
};
export type ListModelsResponse = {
    object?: string;
    data?: Model[];
    has_more?: boolean;
    total_items?: number;
};
export type GetModelRequest = {
    id?: string;
    include_loading_model?: boolean;
};
export type DeleteModelRequest = {
    id?: string;
};
export type DeleteModelResponse = {
    id?: string;
    object?: string;
    deleted?: boolean;
};
export type ActivateModelRequest = {
    id?: string;
};
export type ActivateModelResponse = {};
export type DeactivateModelRequest = {
    id?: string;
};
export type DeactivateModelResponse = {};
export type StorageConfig = {
    path_prefix?: string;
};
export type CreateStorageConfigRequest = {
    path_prefix?: string;
};
export type GetStorageConfigRequest = {};
export type RegisterModelRequest = {
    id?: string;
    base_model?: string;
    suffix?: string;
    organization_id?: string;
    project_id?: string;
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
    base_model?: string;
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
    gguf_model_path?: string;
    source_repository?: SourceRepository;
};
export type BaseModel = {
    id?: string;
    created?: string;
    object?: string;
};
export type GetBaseModelPathRequest = {
    id?: string;
};
export type GetBaseModelPathResponse = {
    formats?: ModelFormat[];
    path?: string;
    gguf_model_path?: string;
};
export type CreateHFModelRepoRequest = {
    name?: string;
};
export type HFModelRepo = {
    name?: string;
};
export type GetHFModelRepoRequest = {
    name?: string;
};
export type AcquireUnloadedBaseModelRequest = {};
export type AcquireUnloadedBaseModelResponse = {
    base_model_id?: string;
    source_repository?: SourceRepository;
};
export type UpdateBaseModelLoadingStatusRequestSuccess = {};
export type UpdateBaseModelLoadingStatusRequestFailure = {
    reason?: string;
};
type BaseUpdateBaseModelLoadingStatusRequest = {
    id?: string;
};
export type UpdateBaseModelLoadingStatusRequest = BaseUpdateBaseModelLoadingStatusRequest & OneOf<{
    success: UpdateBaseModelLoadingStatusRequestSuccess;
    failure: UpdateBaseModelLoadingStatusRequestFailure;
}>;
export type UpdateBaseModelLoadingStatusResponse = {};
export type AcquireUnloadedModelRequest = {};
export type AcquireUnloadedModelResponse = {
    model_id?: string;
    is_base_model?: boolean;
    source_repository?: SourceRepository;
    model_file_location?: string;
    dest_path?: string;
};
export type UpdateModelLoadingStatusRequestSuccess = {};
export type UpdateModelLoadingStatusRequestFailure = {
    reason?: string;
};
type BaseUpdateModelLoadingStatusRequest = {
    id?: string;
    is_base_model?: boolean;
};
export type UpdateModelLoadingStatusRequest = BaseUpdateModelLoadingStatusRequest & OneOf<{
    success: UpdateModelLoadingStatusRequestSuccess;
    failure: UpdateModelLoadingStatusRequestFailure;
}>;
export type UpdateModelLoadingStatusResponse = {};
export declare class ModelsService {
    static ListModels(req: ListModelsRequest, initReq?: fm.InitReq): Promise<ListModelsResponse>;
    static GetModel(req: GetModelRequest, initReq?: fm.InitReq): Promise<Model>;
    static DeleteModel(req: DeleteModelRequest, initReq?: fm.InitReq): Promise<DeleteModelResponse>;
    static CreateModel(req: CreateModelRequest, initReq?: fm.InitReq): Promise<Model>;
    static ActivateModel(req: ActivateModelRequest, initReq?: fm.InitReq): Promise<ActivateModelResponse>;
    static DeactivateModel(req: DeactivateModelRequest, initReq?: fm.InitReq): Promise<DeactivateModelResponse>;
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
    static CreateHFModelRepo(req: CreateHFModelRepoRequest, initReq?: fm.InitReq): Promise<HFModelRepo>;
    static GetHFModelRepo(req: GetHFModelRepoRequest, initReq?: fm.InitReq): Promise<HFModelRepo>;
    static AcquireUnloadedBaseModel(req: AcquireUnloadedBaseModelRequest, initReq?: fm.InitReq): Promise<AcquireUnloadedBaseModelResponse>;
    static AcquireUnloadedModel(req: AcquireUnloadedModelRequest, initReq?: fm.InitReq): Promise<AcquireUnloadedModelResponse>;
    static UpdateBaseModelLoadingStatus(req: UpdateBaseModelLoadingStatusRequest, initReq?: fm.InitReq): Promise<UpdateBaseModelLoadingStatusResponse>;
    static UpdateModelLoadingStatus(req: UpdateModelLoadingStatusRequest, initReq?: fm.InitReq): Promise<UpdateModelLoadingStatusResponse>;
}
export {};
