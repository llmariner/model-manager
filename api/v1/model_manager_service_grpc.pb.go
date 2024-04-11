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
}

type modelsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModelsServiceClient(cc grpc.ClientConnInterface) ModelsServiceClient {
	return &modelsServiceClient{cc}
}

func (c *modelsServiceClient) ListModels(ctx context.Context, in *ListModelsRequest, opts ...grpc.CallOption) (*ListModelsResponse, error) {
	out := new(ListModelsResponse)
	err := c.cc.Invoke(ctx, "/llmoperator.models.server.v1.ModelsService/ListModels", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsServiceClient) GetModel(ctx context.Context, in *GetModelRequest, opts ...grpc.CallOption) (*Model, error) {
	out := new(Model)
	err := c.cc.Invoke(ctx, "/llmoperator.models.server.v1.ModelsService/GetModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsServiceClient) DeleteModel(ctx context.Context, in *DeleteModelRequest, opts ...grpc.CallOption) (*DeleteModelResponse, error) {
	out := new(DeleteModelResponse)
	err := c.cc.Invoke(ctx, "/llmoperator.models.server.v1.ModelsService/DeleteModel", in, out, opts...)
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
		FullMethod: "/llmoperator.models.server.v1.ModelsService/ListModels",
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
		FullMethod: "/llmoperator.models.server.v1.ModelsService/GetModel",
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
		FullMethod: "/llmoperator.models.server.v1.ModelsService/DeleteModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsServiceServer).DeleteModel(ctx, req.(*DeleteModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ModelsService_ServiceDesc is the grpc.ServiceDesc for ModelsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ModelsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "llmoperator.models.server.v1.ModelsService",
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
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/model_manager_service.proto",
}

// ModelsInternalServiceClient is the client API for ModelsInternalService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ModelsInternalServiceClient interface {
	RegisterModel(ctx context.Context, in *RegisterModelRequest, opts ...grpc.CallOption) (*RegisterModelResponse, error)
	PublishModel(ctx context.Context, in *PublishModelRequest, opts ...grpc.CallOption) (*PublishModelResponse, error)
}

type modelsInternalServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModelsInternalServiceClient(cc grpc.ClientConnInterface) ModelsInternalServiceClient {
	return &modelsInternalServiceClient{cc}
}

func (c *modelsInternalServiceClient) RegisterModel(ctx context.Context, in *RegisterModelRequest, opts ...grpc.CallOption) (*RegisterModelResponse, error) {
	out := new(RegisterModelResponse)
	err := c.cc.Invoke(ctx, "/llmoperator.models.server.v1.ModelsInternalService/RegisterModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *modelsInternalServiceClient) PublishModel(ctx context.Context, in *PublishModelRequest, opts ...grpc.CallOption) (*PublishModelResponse, error) {
	out := new(PublishModelResponse)
	err := c.cc.Invoke(ctx, "/llmoperator.models.server.v1.ModelsInternalService/PublishModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ModelsInternalServiceServer is the server API for ModelsInternalService service.
// All implementations must embed UnimplementedModelsInternalServiceServer
// for forward compatibility
type ModelsInternalServiceServer interface {
	RegisterModel(context.Context, *RegisterModelRequest) (*RegisterModelResponse, error)
	PublishModel(context.Context, *PublishModelRequest) (*PublishModelResponse, error)
	mustEmbedUnimplementedModelsInternalServiceServer()
}

// UnimplementedModelsInternalServiceServer must be embedded to have forward compatible implementations.
type UnimplementedModelsInternalServiceServer struct {
}

func (UnimplementedModelsInternalServiceServer) RegisterModel(context.Context, *RegisterModelRequest) (*RegisterModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterModel not implemented")
}
func (UnimplementedModelsInternalServiceServer) PublishModel(context.Context, *PublishModelRequest) (*PublishModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PublishModel not implemented")
}
func (UnimplementedModelsInternalServiceServer) mustEmbedUnimplementedModelsInternalServiceServer() {}

// UnsafeModelsInternalServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ModelsInternalServiceServer will
// result in compilation errors.
type UnsafeModelsInternalServiceServer interface {
	mustEmbedUnimplementedModelsInternalServiceServer()
}

func RegisterModelsInternalServiceServer(s grpc.ServiceRegistrar, srv ModelsInternalServiceServer) {
	s.RegisterService(&ModelsInternalService_ServiceDesc, srv)
}

func _ModelsInternalService_RegisterModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsInternalServiceServer).RegisterModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmoperator.models.server.v1.ModelsInternalService/RegisterModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsInternalServiceServer).RegisterModel(ctx, req.(*RegisterModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModelsInternalService_PublishModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelsInternalServiceServer).PublishModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmoperator.models.server.v1.ModelsInternalService/PublishModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelsInternalServiceServer).PublishModel(ctx, req.(*PublishModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ModelsInternalService_ServiceDesc is the grpc.ServiceDesc for ModelsInternalService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ModelsInternalService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "llmoperator.models.server.v1.ModelsInternalService",
	HandlerType: (*ModelsInternalServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterModel",
			Handler:    _ModelsInternalService_RegisterModel_Handler,
		},
		{
			MethodName: "PublishModel",
			Handler:    _ModelsInternalService_PublishModel_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/model_manager_service.proto",
}
