// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: server.proto

package interfaceRetrieverServer

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

const (
	DataRetrieval_RetrieveFramesAndData_FullMethodName          = "/interfaceRetrieverServer.DataRetrieval/RetrieveFramesAndData"
	DataRetrieval_RetrieveLastConfirmDataStoreId_FullMethodName = "/interfaceRetrieverServer.DataRetrieval/RetrieveLastConfirmDataStoreId"
)

// DataRetrievalClient is the client API for DataRetrieval service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataRetrievalClient interface {
	RetrieveFramesAndData(ctx context.Context, in *FramesAndDataRequest, opts ...grpc.CallOption) (*FramesAndDataReply, error)
	RetrieveLastConfirmDataStoreId(ctx context.Context, in *LastDataStoreIdRequest, opts ...grpc.CallOption) (*LastDataStoreIdReply, error)
}

type dataRetrievalClient struct {
	cc grpc.ClientConnInterface
}

func NewDataRetrievalClient(cc grpc.ClientConnInterface) DataRetrievalClient {
	return &dataRetrievalClient{cc}
}

func (c *dataRetrievalClient) RetrieveFramesAndData(ctx context.Context, in *FramesAndDataRequest, opts ...grpc.CallOption) (*FramesAndDataReply, error) {
	out := new(FramesAndDataReply)
	err := c.cc.Invoke(ctx, DataRetrieval_RetrieveFramesAndData_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataRetrievalClient) RetrieveLastConfirmDataStoreId(ctx context.Context, in *LastDataStoreIdRequest, opts ...grpc.CallOption) (*LastDataStoreIdReply, error) {
	out := new(LastDataStoreIdReply)
	err := c.cc.Invoke(ctx, DataRetrieval_RetrieveLastConfirmDataStoreId_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataRetrievalServer is the server API for DataRetrieval service.
// All implementations must embed UnimplementedDataRetrievalServer
// for forward compatibility
type DataRetrievalServer interface {
	RetrieveFramesAndData(context.Context, *FramesAndDataRequest) (*FramesAndDataReply, error)
	RetrieveLastConfirmDataStoreId(context.Context, *LastDataStoreIdRequest) (*LastDataStoreIdReply, error)
	mustEmbedUnimplementedDataRetrievalServer()
}

// UnimplementedDataRetrievalServer must be embedded to have forward compatible implementations.
type UnimplementedDataRetrievalServer struct {
}

func (UnimplementedDataRetrievalServer) RetrieveFramesAndData(context.Context, *FramesAndDataRequest) (*FramesAndDataReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveFramesAndData not implemented")
}
func (UnimplementedDataRetrievalServer) RetrieveLastConfirmDataStoreId(context.Context, *LastDataStoreIdRequest) (*LastDataStoreIdReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveLastConfirmDataStoreId not implemented")
}
func (UnimplementedDataRetrievalServer) mustEmbedUnimplementedDataRetrievalServer() {}

// UnsafeDataRetrievalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataRetrievalServer will
// result in compilation errors.
type UnsafeDataRetrievalServer interface {
	mustEmbedUnimplementedDataRetrievalServer()
}

func RegisterDataRetrievalServer(s grpc.ServiceRegistrar, srv DataRetrievalServer) {
	s.RegisterService(&DataRetrieval_ServiceDesc, srv)
}

func _DataRetrieval_RetrieveFramesAndData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FramesAndDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataRetrievalServer).RetrieveFramesAndData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataRetrieval_RetrieveFramesAndData_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataRetrievalServer).RetrieveFramesAndData(ctx, req.(*FramesAndDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataRetrieval_RetrieveLastConfirmDataStoreId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LastDataStoreIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataRetrievalServer).RetrieveLastConfirmDataStoreId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataRetrieval_RetrieveLastConfirmDataStoreId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataRetrievalServer).RetrieveLastConfirmDataStoreId(ctx, req.(*LastDataStoreIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DataRetrieval_ServiceDesc is the grpc.ServiceDesc for DataRetrieval service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataRetrieval_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "interfaceRetrieverServer.DataRetrieval",
	HandlerType: (*DataRetrievalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RetrieveFramesAndData",
			Handler:    _DataRetrieval_RetrieveFramesAndData_Handler,
		},
		{
			MethodName: "RetrieveLastConfirmDataStoreId",
			Handler:    _DataRetrieval_RetrieveLastConfirmDataStoreId_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}