// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ModelsServiceClient is the client API for ModelsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ModelsServiceClient interface {
	ListModels(ctx context.Context, in *ListModelsRequest, opts ...grpc.CallOption) (*ListModelsResponse, error)
	GetModel(ctx context.Context, in *GetModelRequest, opts ...grpc.CallOption) (*Model, error)
	DeleteModel(ctx context.Context, in *DeleteModelRequest, opts ...grpc.CallOption) (*DeleteModelResponse, error)
	// The following API endpoints are not part of the OpenAPI API specification.
	ListBaseModels(ctx context.Context, in *ListBaseModelsRequest, opts ...grpc.CallOption) (*ListBaseModelsResponse, error)
}

type modelsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModelsServiceClient(cc grpc.ClientConnInterface) ModelsServiceClient {
	return &modelsServiceClient{cc}
}

func (c *modelsServiceClient) ListModels(ctx context.Context, in *ListModelsRequest, opts ...grpc.CallOption) (*ListModelsResponse, error) {
	out := new(ListModelsResponse)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsService/ListModels", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsServiceClient) GetModel(ctx context.Context, in *GetModelRequest, opts ...grpc.CallOption) (*Model, error) {
	out := new(Model)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsService/GetModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsServiceClient) DeleteModel(ctx context.Context, in *DeleteModelRequest, opts ...grpc.CallOption) (*DeleteModelResponse, error) {
	out := new(DeleteModelResponse)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsService/DeleteModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsServiceClient) ListBaseModels(ctx context.Context, in *ListBaseModelsRequest, opts ...grpc.CallOption) (*ListBaseModelsResponse, error) {
	out := new(ListBaseModelsResponse)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsService/ListBaseModels", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ModelsServiceServer is the server API for ModelsService service.
// All implementations must embed UnimplementedModelsServiceServer
// for forward compatibility
type ModelsServiceServer interface {
	ListModels(context.Context, *ListModelsRequest) (*ListModelsResponse, error)
	GetModel(context.Context, *GetModelRequest) (*Model, error)
	DeleteModel(context.Context, *DeleteModelRequest) (*DeleteModelResponse, error)
	// The following API endpoints are not part of the OpenAPI API specification.
	ListBaseModels(context.Context, *ListBaseModelsRequest) (*ListBaseModelsResponse, error)
	mustEmbedUnimplementedModelsServiceServer()
}

// UnimplementedModelsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedModelsServiceServer struct {
}

func (UnimplementedModelsServiceServer) ListModels(context.Context, *ListModelsRequest) (*ListModelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListModels not implemented")
}
func (UnimplementedModelsServiceServer) GetModel(context.Context, *GetModelRequest) (*Model, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModel not implemented")
}
func (UnimplementedModelsServiceServer) DeleteModel(context.Context, *DeleteModelRequest) (*DeleteModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteModel not implemented")
}
func (UnimplementedModelsServiceServer) ListBaseModels(context.Context, *ListBaseModelsRequest) (*ListBaseModelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListBaseModels not implemented")
}
func (UnimplementedModelsServiceServer) mustEmbedUnimplementedModelsServiceServer() {}

// UnsafeModelsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ModelsServiceServer will
// result in compilation errors.
type UnsafeModelsServiceServer interface {
	mustEmbedUnimplementedModelsServiceServer()
}

func RegisterModelsServiceServer(s grpc.ServiceRegistrar, srv ModelsServiceServer) {
	s.RegisterService(&ModelsService_ServiceDesc, srv)
}

func _ModelsService_ListModels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListModelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsServiceServer).ListModels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsService/ListModels",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsServiceServer).ListModels(ctx, req.(*ListModelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsService_GetModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsServiceServer).GetModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsService/GetModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsServiceServer).GetModel(ctx, req.(*GetModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsService_DeleteModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsServiceServer).DeleteModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsService/DeleteModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsServiceServer).DeleteModel(ctx, req.(*DeleteModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsService_ListBaseModels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListBaseModelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsServiceServer).ListBaseModels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsService/ListBaseModels",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsServiceServer).ListBaseModels(ctx, req.(*ListBaseModelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ModelsService_ServiceDesc is the grpc.ServiceDesc for ModelsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ModelsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "llmariner.models.server.v1.ModelsService",
	HandlerType: (*ModelsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListModels",
			Handler:    _ModelsService_ListModels_Handler,
		},
		{
			MethodName: "GetModel",
			Handler:    _ModelsService_GetModel_Handler,
		},
		{
			MethodName: "DeleteModel",
			Handler:    _ModelsService_DeleteModel_Handler,
		},
		{
			MethodName: "ListBaseModels",
			Handler:    _ModelsService_ListBaseModels_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/model_manager_service.proto",
}

// ModelsWorkerServiceClient is the client API for ModelsWorkerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ModelsWorkerServiceClient interface {
	// CreateStorageConfig creates a new storage config. Used by model-manager-loader.
	CreateStorageConfig(ctx context.Context, in *CreateStorageConfigRequest, opts ...grpc.CallOption) (*StorageConfig, error)
	// GetStorageConfig gets a storage config. Used by model-manager-loader.
	GetStorageConfig(ctx context.Context, in *GetStorageConfigRequest, opts ...grpc.CallOption) (*StorageConfig, error)
	// GetModel gets a model. Used by inference-manager-engine.
	GetModel(ctx context.Context, in *GetModelRequest, opts ...grpc.CallOption) (*Model, error)
	// RegisterModel registers a new fine-tuned model. Used by job-manager-dispatcher and model-manager-loader.
	// The model is not published until PublishModel is called.
	RegisterModel(ctx context.Context, in *RegisterModelRequest, opts ...grpc.CallOption) (*RegisterModelResponse, error)
	// PublishModel publishes the fine-tuned model. Used by job-manager-dispatcher and model-manager-loader.
	PublishModel(ctx context.Context, in *PublishModelRequest, opts ...grpc.CallOption) (*PublishModelResponse, error)
	// GetModelPath returns the path of the model. Used by inference-manager-engine and model-manager-loader.
	GetModelPath(ctx context.Context, in *GetModelPathRequest, opts ...grpc.CallOption) (*GetModelPathResponse, error)
	// GetModelAttributes returns the attributes of the model. Used by inference-manager-engine.
	GetModelAttributes(ctx context.Context, in *GetModelAttributesRequest, opts ...grpc.CallOption) (*ModelAttributes, error)
	// CreateBaseModel creates a new base model. Used by model-manager-loader.
	CreateBaseModel(ctx context.Context, in *CreateBaseModelRequest, opts ...grpc.CallOption) (*BaseModel, error)
	// GetBaseModelPath returns the path of the base model. Used by job-manager-dispatcher,
	// inference-manager-engine, and model-manager-loader.
	GetBaseModelPath(ctx context.Context, in *GetBaseModelPathRequest, opts ...grpc.CallOption) (*GetBaseModelPathResponse, error)
}

type modelsWorkerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModelsWorkerServiceClient(cc grpc.ClientConnInterface) ModelsWorkerServiceClient {
	return &modelsWorkerServiceClient{cc}
}

func (c *modelsWorkerServiceClient) CreateStorageConfig(ctx context.Context, in *CreateStorageConfigRequest, opts ...grpc.CallOption) (*StorageConfig, error) {
	out := new(StorageConfig)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/CreateStorageConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) GetStorageConfig(ctx context.Context, in *GetStorageConfigRequest, opts ...grpc.CallOption) (*StorageConfig, error) {
	out := new(StorageConfig)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/GetStorageConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) GetModel(ctx context.Context, in *GetModelRequest, opts ...grpc.CallOption) (*Model, error) {
	out := new(Model)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/GetModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) RegisterModel(ctx context.Context, in *RegisterModelRequest, opts ...grpc.CallOption) (*RegisterModelResponse, error) {
	out := new(RegisterModelResponse)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/RegisterModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) PublishModel(ctx context.Context, in *PublishModelRequest, opts ...grpc.CallOption) (*PublishModelResponse, error) {
	out := new(PublishModelResponse)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/PublishModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) GetModelPath(ctx context.Context, in *GetModelPathRequest, opts ...grpc.CallOption) (*GetModelPathResponse, error) {
	out := new(GetModelPathResponse)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/GetModelPath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) GetModelAttributes(ctx context.Context, in *GetModelAttributesRequest, opts ...grpc.CallOption) (*ModelAttributes, error) {
	out := new(ModelAttributes)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/GetModelAttributes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) CreateBaseModel(ctx context.Context, in *CreateBaseModelRequest, opts ...grpc.CallOption) (*BaseModel, error) {
	out := new(BaseModel)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/CreateBaseModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsWorkerServiceClient) GetBaseModelPath(ctx context.Context, in *GetBaseModelPathRequest, opts ...grpc.CallOption) (*GetBaseModelPathResponse, error) {
	out := new(GetBaseModelPathResponse)
	err := c.cc.Invoke(ctx, "/llmariner.models.server.v1.ModelsWorkerService/GetBaseModelPath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ModelsWorkerServiceServer is the server API for ModelsWorkerService service.
// All implementations must embed UnimplementedModelsWorkerServiceServer
// for forward compatibility
type ModelsWorkerServiceServer interface {
	// CreateStorageConfig creates a new storage config. Used by model-manager-loader.
	CreateStorageConfig(context.Context, *CreateStorageConfigRequest) (*StorageConfig, error)
	// GetStorageConfig gets a storage config. Used by model-manager-loader.
	GetStorageConfig(context.Context, *GetStorageConfigRequest) (*StorageConfig, error)
	// GetModel gets a model. Used by inference-manager-engine.
	GetModel(context.Context, *GetModelRequest) (*Model, error)
	// RegisterModel registers a new fine-tuned model. Used by job-manager-dispatcher and model-manager-loader.
	// The model is not published until PublishModel is called.
	RegisterModel(context.Context, *RegisterModelRequest) (*RegisterModelResponse, error)
	// PublishModel publishes the fine-tuned model. Used by job-manager-dispatcher and model-manager-loader.
	PublishModel(context.Context, *PublishModelRequest) (*PublishModelResponse, error)
	// GetModelPath returns the path of the model. Used by inference-manager-engine and model-manager-loader.
	GetModelPath(context.Context, *GetModelPathRequest) (*GetModelPathResponse, error)
	// GetModelAttributes returns the attributes of the model. Used by inference-manager-engine.
	GetModelAttributes(context.Context, *GetModelAttributesRequest) (*ModelAttributes, error)
	// CreateBaseModel creates a new base model. Used by model-manager-loader.
	CreateBaseModel(context.Context, *CreateBaseModelRequest) (*BaseModel, error)
	// GetBaseModelPath returns the path of the base model. Used by job-manager-dispatcher,
	// inference-manager-engine, and model-manager-loader.
	GetBaseModelPath(context.Context, *GetBaseModelPathRequest) (*GetBaseModelPathResponse, error)
	mustEmbedUnimplementedModelsWorkerServiceServer()
}

// UnimplementedModelsWorkerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedModelsWorkerServiceServer struct {
}

func (UnimplementedModelsWorkerServiceServer) CreateStorageConfig(context.Context, *CreateStorageConfigRequest) (*StorageConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateStorageConfig not implemented")
}
func (UnimplementedModelsWorkerServiceServer) GetStorageConfig(context.Context, *GetStorageConfigRequest) (*StorageConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStorageConfig not implemented")
}
func (UnimplementedModelsWorkerServiceServer) GetModel(context.Context, *GetModelRequest) (*Model, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModel not implemented")
}
func (UnimplementedModelsWorkerServiceServer) RegisterModel(context.Context, *RegisterModelRequest) (*RegisterModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterModel not implemented")
}
func (UnimplementedModelsWorkerServiceServer) PublishModel(context.Context, *PublishModelRequest) (*PublishModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PublishModel not implemented")
}
func (UnimplementedModelsWorkerServiceServer) GetModelPath(context.Context, *GetModelPathRequest) (*GetModelPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModelPath not implemented")
}
func (UnimplementedModelsWorkerServiceServer) GetModelAttributes(context.Context, *GetModelAttributesRequest) (*ModelAttributes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModelAttributes not implemented")
}
func (UnimplementedModelsWorkerServiceServer) CreateBaseModel(context.Context, *CreateBaseModelRequest) (*BaseModel, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateBaseModel not implemented")
}
func (UnimplementedModelsWorkerServiceServer) GetBaseModelPath(context.Context, *GetBaseModelPathRequest) (*GetBaseModelPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBaseModelPath not implemented")
}
func (UnimplementedModelsWorkerServiceServer) mustEmbedUnimplementedModelsWorkerServiceServer() {}

// UnsafeModelsWorkerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ModelsWorkerServiceServer will
// result in compilation errors.
type UnsafeModelsWorkerServiceServer interface {
	mustEmbedUnimplementedModelsWorkerServiceServer()
}

func RegisterModelsWorkerServiceServer(s grpc.ServiceRegistrar, srv ModelsWorkerServiceServer) {
	s.RegisterService(&ModelsWorkerService_ServiceDesc, srv)
}

func _ModelsWorkerService_CreateStorageConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateStorageConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).CreateStorageConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/CreateStorageConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).CreateStorageConfig(ctx, req.(*CreateStorageConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_GetStorageConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStorageConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).GetStorageConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/GetStorageConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).GetStorageConfig(ctx, req.(*GetStorageConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_GetModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).GetModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/GetModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).GetModel(ctx, req.(*GetModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_RegisterModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).RegisterModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/RegisterModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).RegisterModel(ctx, req.(*RegisterModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_PublishModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).PublishModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/PublishModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).PublishModel(ctx, req.(*PublishModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_GetModelPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModelPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).GetModelPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/GetModelPath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).GetModelPath(ctx, req.(*GetModelPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_GetModelAttributes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModelAttributesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).GetModelAttributes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/GetModelAttributes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).GetModelAttributes(ctx, req.(*GetModelAttributesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_CreateBaseModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateBaseModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).CreateBaseModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/CreateBaseModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).CreateBaseModel(ctx, req.(*CreateBaseModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsWorkerService_GetBaseModelPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBaseModelPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsWorkerServiceServer).GetBaseModelPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmariner.models.server.v1.ModelsWorkerService/GetBaseModelPath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsWorkerServiceServer).GetBaseModelPath(ctx, req.(*GetBaseModelPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ModelsWorkerService_ServiceDesc is the grpc.ServiceDesc for ModelsWorkerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ModelsWorkerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "llmariner.models.server.v1.ModelsWorkerService",
	HandlerType: (*ModelsWorkerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateStorageConfig",
			Handler:    _ModelsWorkerService_CreateStorageConfig_Handler,
		},
		{
			MethodName: "GetStorageConfig",
			Handler:    _ModelsWorkerService_GetStorageConfig_Handler,
		},
		{
			MethodName: "GetModel",
			Handler:    _ModelsWorkerService_GetModel_Handler,
		},
		{
			MethodName: "RegisterModel",
			Handler:    _ModelsWorkerService_RegisterModel_Handler,
		},
		{
			MethodName: "PublishModel",
			Handler:    _ModelsWorkerService_PublishModel_Handler,
		},
		{
			MethodName: "GetModelPath",
			Handler:    _ModelsWorkerService_GetModelPath_Handler,
		},
		{
			MethodName: "GetModelAttributes",
			Handler:    _ModelsWorkerService_GetModelAttributes_Handler,
		},
		{
			MethodName: "CreateBaseModel",
			Handler:    _ModelsWorkerService_CreateBaseModel_Handler,
		},
		{
			MethodName: "GetBaseModelPath",
			Handler:    _ModelsWorkerService_GetBaseModelPath_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/model_manager_service.proto",
}
