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
    ModelFormat["MODEL_FORMAT_OLLAMA"] = "MODEL_FORMAT_OLLAMA";
})(ModelFormat || (ModelFormat = {}));
export var ModelLoadingStatus;
(function (ModelLoadingStatus) {
    ModelLoadingStatus["MODEL_LOADING_STATUS_UNSPECIFIED"] = "MODEL_LOADING_STATUS_UNSPECIFIED";
    ModelLoadingStatus["MODEL_LOADING_STATUS_REQUESTED"] = "MODEL_LOADING_STATUS_REQUESTED";
    ModelLoadingStatus["MODEL_LOADING_STATUS_LOADING"] = "MODEL_LOADING_STATUS_LOADING";
    ModelLoadingStatus["MODEL_LOADING_STATUS_SUCCEEDED"] = "MODEL_LOADING_STATUS_SUCCEEDED";
    ModelLoadingStatus["MODEL_LOADING_STATUS_FAILED"] = "MODEL_LOADING_STATUS_FAILED";
})(ModelLoadingStatus || (ModelLoadingStatus = {}));
export var SourceRepository;
(function (SourceRepository) {
    SourceRepository["SOURCE_REPOSITORY_UNSPECIFIED"] = "SOURCE_REPOSITORY_UNSPECIFIED";
    SourceRepository["SOURCE_REPOSITORY_OBJECT_STORE"] = "SOURCE_REPOSITORY_OBJECT_STORE";
    SourceRepository["SOURCE_REPOSITORY_HUGGING_FACE"] = "SOURCE_REPOSITORY_HUGGING_FACE";
    SourceRepository["SOURCE_REPOSITORY_OLLAMA"] = "SOURCE_REPOSITORY_OLLAMA";
    SourceRepository["SOURCE_REPOSITORY_FINE_TUNING"] = "SOURCE_REPOSITORY_FINE_TUNING";
})(SourceRepository || (SourceRepository = {}));
export var ActivationStatus;
(function (ActivationStatus) {
    ActivationStatus["ACTIVATION_STATUS_UNSPECIFIED"] = "ACTIVATION_STATUS_UNSPECIFIED";
    ActivationStatus["ACTIVATION_STATUS_ACTIVE"] = "ACTIVATION_STATUS_ACTIVE";
    ActivationStatus["ACTIVATION_STATUS_INACTIVE"] = "ACTIVATION_STATUS_INACTIVE";
})(ActivationStatus || (ActivationStatus = {}));
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
    static GetModel(req, initReq) {
        return fm.fetchReq(`/v1/models/${req["id=**"]}?${fm.renderURLSearchParams(req, ["id=**"])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static ListModels(req, initReq) {
        return fm.fetchReq(`/v1/models?${fm.renderURLSearchParams(req, [])}`, Object.assign(Object.assign({}, initReq), { method: "GET" }));
    }
    static DeleteModel(req, initReq) {
        return fm.fetchReq(`/v1/models/${req["id=**"]}`, Object.assign(Object.assign({}, initReq), { method: "DELETE" }));
    }
    static CreateModel(req, initReq) {
        return fm.fetchReq(`/v1/models`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static UpdateModel(req, initReq) {
        return fm.fetchReq(`/v1/models/${req["model.id"]}`, Object.assign(Object.assign({}, initReq), { method: "PATCH", body: JSON.stringify(req) }));
    }
    static ActivateModel(req, initReq) {
        return fm.fetchReq(`/v1/models/${req["id"]}:activate`, Object.assign(Object.assign({}, initReq), { method: "POST" }));
    }
    static DeactivateModel(req, initReq) {
        return fm.fetchReq(`/v1/models/${req["id"]}:deactivate`, Object.assign(Object.assign({}, initReq), { method: "POST" }));
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
    static ListModels(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/ListModels`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
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
    static AcquireUnloadedBaseModel(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/AcquireUnloadedBaseModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static AcquireUnloadedModel(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/AcquireUnloadedModel`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static UpdateBaseModelLoadingStatus(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/UpdateBaseModelLoadingStatus`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
    static UpdateModelLoadingStatus(req, initReq) {
        return fm.fetchReq(`/llmariner.models.server.v1.ModelsWorkerService/UpdateModelLoadingStatus`, Object.assign(Object.assign({}, initReq), { method: "POST", body: JSON.stringify(req) }));
    }
}
