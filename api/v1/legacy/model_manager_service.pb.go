// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        (unknown)
// source: api/v1/legacy/model_manager_service.proto

package legacy

import (
	v1 "github.com/llmariner/model-manager/api/v1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_api_v1_legacy_model_manager_service_proto protoreflect.FileDescriptor

var file_api_v1_legacy_model_manager_service_proto_rawDesc = []byte{
	0x0a, 0x29, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x65, 0x67, 0x61, 0x63, 0x79, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x6c, 0x6c, 0x6d,
	0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xb4, 0x08, 0x0a, 0x13,
	0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x7a, 0x0a, 0x13, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x36, 0x2e, 0x6c, 0x6c, 0x6d,
	0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x29, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x00, 0x12,
	0x74, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x33, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72,
	0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x22, 0x00, 0x12, 0x5c, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x12, 0x2b, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21,
	0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x22, 0x00, 0x12, 0x76, 0x0a, 0x0d, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4d,
	0x6f, 0x64, 0x65, 0x6c, 0x12, 0x30, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e,
	0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x73, 0x0a, 0x0c, 0x50,
	0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x2f, 0x2e, 0x6c, 0x6c,
	0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68,
	0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x30, 0x2e, 0x6c,
	0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73,
	0x68, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x73, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x50, 0x61, 0x74, 0x68,
	0x12, 0x2f, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x50, 0x61, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x30, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x50, 0x61, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x7a, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x35, 0x2e, 0x6c, 0x6c,
	0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2b, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x22,
	0x00, 0x12, 0x6e, 0x0a, 0x0f, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x42, 0x61, 0x73, 0x65, 0x4d,
	0x6f, 0x64, 0x65, 0x6c, 0x12, 0x32, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x42, 0x61, 0x73, 0x65, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72,
	0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x61, 0x73, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x22,
	0x00, 0x12, 0x7f, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x42, 0x61, 0x73, 0x65, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x50, 0x61, 0x74, 0x68, 0x12, 0x33, 0x2e, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65,
	0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x61, 0x73, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x50,
	0x61, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x34, 0x2e, 0x6c, 0x6c, 0x6d,
	0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x61, 0x73, 0x65, 0x4d,
	0x6f, 0x64, 0x65, 0x6c, 0x50, 0x61, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6c, 0x6c, 0x6d, 0x61, 0x72, 0x69, 0x6e, 0x65, 0x72, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x2d, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f,
	0x6c, 0x65, 0x67, 0x61, 0x63, 0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_api_v1_legacy_model_manager_service_proto_goTypes = []interface{}{
	(*v1.CreateStorageConfigRequest)(nil), // 0: llmariner.models.server.v1.CreateStorageConfigRequest
	(*v1.GetStorageConfigRequest)(nil),    // 1: llmariner.models.server.v1.GetStorageConfigRequest
	(*v1.GetModelRequest)(nil),            // 2: llmariner.models.server.v1.GetModelRequest
	(*v1.RegisterModelRequest)(nil),       // 3: llmariner.models.server.v1.RegisterModelRequest
	(*v1.PublishModelRequest)(nil),        // 4: llmariner.models.server.v1.PublishModelRequest
	(*v1.GetModelPathRequest)(nil),        // 5: llmariner.models.server.v1.GetModelPathRequest
	(*v1.GetModelAttributesRequest)(nil),  // 6: llmariner.models.server.v1.GetModelAttributesRequest
	(*v1.CreateBaseModelRequest)(nil),     // 7: llmariner.models.server.v1.CreateBaseModelRequest
	(*v1.GetBaseModelPathRequest)(nil),    // 8: llmariner.models.server.v1.GetBaseModelPathRequest
	(*v1.StorageConfig)(nil),              // 9: llmariner.models.server.v1.StorageConfig
	(*v1.Model)(nil),                      // 10: llmariner.models.server.v1.Model
	(*v1.RegisterModelResponse)(nil),      // 11: llmariner.models.server.v1.RegisterModelResponse
	(*v1.PublishModelResponse)(nil),       // 12: llmariner.models.server.v1.PublishModelResponse
	(*v1.GetModelPathResponse)(nil),       // 13: llmariner.models.server.v1.GetModelPathResponse
	(*v1.ModelAttributes)(nil),            // 14: llmariner.models.server.v1.ModelAttributes
	(*v1.BaseModel)(nil),                  // 15: llmariner.models.server.v1.BaseModel
	(*v1.GetBaseModelPathResponse)(nil),   // 16: llmariner.models.server.v1.GetBaseModelPathResponse
}
var file_api_v1_legacy_model_manager_service_proto_depIdxs = []int32{
	0,  // 0: llmoperator.models.server.v1.ModelsWorkerService.CreateStorageConfig:input_type -> llmariner.models.server.v1.CreateStorageConfigRequest
	1,  // 1: llmoperator.models.server.v1.ModelsWorkerService.GetStorageConfig:input_type -> llmariner.models.server.v1.GetStorageConfigRequest
	2,  // 2: llmoperator.models.server.v1.ModelsWorkerService.GetModel:input_type -> llmariner.models.server.v1.GetModelRequest
	3,  // 3: llmoperator.models.server.v1.ModelsWorkerService.RegisterModel:input_type -> llmariner.models.server.v1.RegisterModelRequest
	4,  // 4: llmoperator.models.server.v1.ModelsWorkerService.PublishModel:input_type -> llmariner.models.server.v1.PublishModelRequest
	5,  // 5: llmoperator.models.server.v1.ModelsWorkerService.GetModelPath:input_type -> llmariner.models.server.v1.GetModelPathRequest
	6,  // 6: llmoperator.models.server.v1.ModelsWorkerService.GetModelAttributes:input_type -> llmariner.models.server.v1.GetModelAttributesRequest
	7,  // 7: llmoperator.models.server.v1.ModelsWorkerService.CreateBaseModel:input_type -> llmariner.models.server.v1.CreateBaseModelRequest
	8,  // 8: llmoperator.models.server.v1.ModelsWorkerService.GetBaseModelPath:input_type -> llmariner.models.server.v1.GetBaseModelPathRequest
	9,  // 9: llmoperator.models.server.v1.ModelsWorkerService.CreateStorageConfig:output_type -> llmariner.models.server.v1.StorageConfig
	9,  // 10: llmoperator.models.server.v1.ModelsWorkerService.GetStorageConfig:output_type -> llmariner.models.server.v1.StorageConfig
	10, // 11: llmoperator.models.server.v1.ModelsWorkerService.GetModel:output_type -> llmariner.models.server.v1.Model
	11, // 12: llmoperator.models.server.v1.ModelsWorkerService.RegisterModel:output_type -> llmariner.models.server.v1.RegisterModelResponse
	12, // 13: llmoperator.models.server.v1.ModelsWorkerService.PublishModel:output_type -> llmariner.models.server.v1.PublishModelResponse
	13, // 14: llmoperator.models.server.v1.ModelsWorkerService.GetModelPath:output_type -> llmariner.models.server.v1.GetModelPathResponse
	14, // 15: llmoperator.models.server.v1.ModelsWorkerService.GetModelAttributes:output_type -> llmariner.models.server.v1.ModelAttributes
	15, // 16: llmoperator.models.server.v1.ModelsWorkerService.CreateBaseModel:output_type -> llmariner.models.server.v1.BaseModel
	16, // 17: llmoperator.models.server.v1.ModelsWorkerService.GetBaseModelPath:output_type -> llmariner.models.server.v1.GetBaseModelPathResponse
	9,  // [9:18] is the sub-list for method output_type
	0,  // [0:9] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_api_v1_legacy_model_manager_service_proto_init() }
func file_api_v1_legacy_model_manager_service_proto_init() {
	if File_api_v1_legacy_model_manager_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_v1_legacy_model_manager_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_v1_legacy_model_manager_service_proto_goTypes,
		DependencyIndexes: file_api_v1_legacy_model_manager_service_proto_depIdxs,
	}.Build()
	File_api_v1_legacy_model_manager_service_proto = out.File
	file_api_v1_legacy_model_manager_service_proto_rawDesc = nil
	file_api_v1_legacy_model_manager_service_proto_goTypes = nil
	file_api_v1_legacy_model_manager_service_proto_depIdxs = nil
}