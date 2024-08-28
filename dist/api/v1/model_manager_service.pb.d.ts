import * as fm from "../../fetch.pb";
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
    baseModel?: string;
    suffix?: string;
    organizationId?: string;
    projectId?: string;
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
export type CreateBaseModelRequest = {
    id?: string;
    path?: string;
    ggufModelPath?: string;
};
export type GetBaseModelPathRequest = {
    id?: string;
};
export type GetBaseModelPathResponse = {
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
    static CreateBaseModel(req: CreateBaseModelRequest, initReq?: fm.InitReq): Promise<BaseModel>;
    static GetBaseModelPath(req: GetBaseModelPathRequest, initReq?: fm.InitReq): Promise<GetBaseModelPathResponse>;
}
