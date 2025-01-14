/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/
import * as fm from "../../fetch.pb";
export var ModelFormat;
(function (ModelFormat) {
    ModelFormat["MODEL_FORMAT_UNSPECIFIED"] = "MODEL_FORMAT_UNSPECIFIED";
    ModelFormat["MODEL_FORMAT_GGUF"] = "MODEL_FORMAT_GGUF";
    ModelFormat["MODEL_FORMAT_HUGGING_FACE"] = "MODEL_FORMAT_HUGGING_FACE";
    ModelFormat["MODEL_FORMAT_NVIDIA_TRITON"] = "MODEL_FORMAT_NVIDIA_TRITON";
})(ModelFormat || (ModelFormat = {}));
export var AdapterType;
(function (AdapterType) {
    AdapterType["ADAPTER_TYPE_UNSPECIFIED"] = "ADAPTER_TYPE_UNSPECIFIED";
    AdapterType["ADAPTER_TYPE_LORA"] = "ADAPTER_TYPE_LORA";
    AdapterType["ADAPTER_TYPE_QLORA"] = "ADAPTER_TYPE_QLORA";
})(AdapterType || (AdapterType = {}));
export var QuantizationType;
(function (QuantizationType) {
    QuantizationType["QUANTIZATION_TYPE_UNSPECIFIED"] = "QUANTIZATION_TYPE_UNSPECIFIED";
    QuantizationType["QUANTIZATION_TYPE_GGUF"] = "QUANTIZATION_TYPE_GGUF";
    QuantizationType["QUANTIZATION_TYPE_AWQ"] = "QUANTIZATION_TYPE_AWQ";
})(QuantizationType || (QuantizationType = {}));
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
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/CreateStorageConfig`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetStorageConfig(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/GetStorageConfig`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetModel(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/GetModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static RegisterModel(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/RegisterModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static PublishModel(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/PublishModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetModelPath(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/GetModelPath`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetModelAttributes(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/GetModelAttributes`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static CreateBaseModel(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/CreateBaseModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetBaseModelPath(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/GetBaseModelPath`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static CreateHFModelRepo(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/CreateHFModelRepo`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static GetHFModelRepo(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/GetHFModelRepo`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
}
