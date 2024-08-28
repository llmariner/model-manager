/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
import * as fm from "../../fetch.pb";
export class ModelsService {
    static ListModels(req, initReq) {
        return fm.fetchReq(`/v1/models?${fm.renderURLSearchParams(req, [])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static GetModel(req, initReq) {
        return fm.fetchReq(`/v1/models/${req["id"]}?${fm.renderURLSearchParams(req, ["id"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteModel(req, initReq) {
        return fm.fetchReq(`/v1/models/${req["id"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static ListBaseModels(req, initReq) {
        return fm.fetchReq(`/v1/basemodels?${fm.renderURLSearchParams(req, [])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
}
export class ModelsWorkerService {
    static CreateStorageConfig(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/CreateStorageConfig`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetStorageConfig(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/GetStorageConfig`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetModel(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/GetModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static RegisterModel(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/RegisterModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static PublishModel(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/PublishModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetModelPath(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/GetModelPath`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static CreateBaseModel(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/CreateBaseModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetBaseModelPath(req, initReq) {
        return fm.fetchReq(`/llmoperator.models.server.v1.ModelsWorkerService/GetBaseModelPath`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
}
