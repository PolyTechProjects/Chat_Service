// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: notification/notification.proto

package notification

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

// NotificationClient is the client API for Notification service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NotificationClient interface {
	NotifyUser(ctx context.Context, in *NotifyUserRequest, opts ...grpc.CallOption) (*NotifyUserResponse, error)
	BindDeviceToUser(ctx context.Context, in *BindDeviceRequest, opts ...grpc.CallOption) (*BindDeviceResponse, error)
	UpdateOldDeviceOnUser(ctx context.Context, in *UpdateOldDeviceRequest, opts ...grpc.CallOption) (*UpdateOldDeviceResponse, error)
	UnbindDeviceFromUser(ctx context.Context, in *UnbindDeviceRequest, opts ...grpc.CallOption) (*UnbindDeviceResponse, error)
	DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error)
}

type notificationClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationClient(cc grpc.ClientConnInterface) NotificationClient {
	return &notificationClient{cc}
}

func (c *notificationClient) NotifyUser(ctx context.Context, in *NotifyUserRequest, opts ...grpc.CallOption) (*NotifyUserResponse, error) {
	out := new(NotifyUserResponse)
	err := c.cc.Invoke(ctx, "/notification.Notification/NotifyUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationClient) BindDeviceToUser(ctx context.Context, in *BindDeviceRequest, opts ...grpc.CallOption) (*BindDeviceResponse, error) {
	out := new(BindDeviceResponse)
	err := c.cc.Invoke(ctx, "/notification.Notification/BindDeviceToUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationClient) UpdateOldDeviceOnUser(ctx context.Context, in *UpdateOldDeviceRequest, opts ...grpc.CallOption) (*UpdateOldDeviceResponse, error) {
	out := new(UpdateOldDeviceResponse)
	err := c.cc.Invoke(ctx, "/notification.Notification/UpdateOldDeviceOnUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationClient) UnbindDeviceFromUser(ctx context.Context, in *UnbindDeviceRequest, opts ...grpc.CallOption) (*UnbindDeviceResponse, error) {
	out := new(UnbindDeviceResponse)
	err := c.cc.Invoke(ctx, "/notification.Notification/UnbindDeviceFromUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationClient) DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error) {
	out := new(DeleteUserResponse)
	err := c.cc.Invoke(ctx, "/notification.Notification/DeleteUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotificationServer is the server API for Notification service.
// All implementations must embed UnimplementedNotificationServer
// for forward compatibility
type NotificationServer interface {
	NotifyUser(context.Context, *NotifyUserRequest) (*NotifyUserResponse, error)
	BindDeviceToUser(context.Context, *BindDeviceRequest) (*BindDeviceResponse, error)
	UpdateOldDeviceOnUser(context.Context, *UpdateOldDeviceRequest) (*UpdateOldDeviceResponse, error)
	UnbindDeviceFromUser(context.Context, *UnbindDeviceRequest) (*UnbindDeviceResponse, error)
	DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error)
	mustEmbedUnimplementedNotificationServer()
}

// UnimplementedNotificationServer must be embedded to have forward compatible implementations.
type UnimplementedNotificationServer struct {
}

func (UnimplementedNotificationServer) NotifyUser(context.Context, *NotifyUserRequest) (*NotifyUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyUser not implemented")
}
func (UnimplementedNotificationServer) BindDeviceToUser(context.Context, *BindDeviceRequest) (*BindDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BindDeviceToUser not implemented")
}
func (UnimplementedNotificationServer) UpdateOldDeviceOnUser(context.Context, *UpdateOldDeviceRequest) (*UpdateOldDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOldDeviceOnUser not implemented")
}
func (UnimplementedNotificationServer) UnbindDeviceFromUser(context.Context, *UnbindDeviceRequest) (*UnbindDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnbindDeviceFromUser not implemented")
}
func (UnimplementedNotificationServer) DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedNotificationServer) mustEmbedUnimplementedNotificationServer() {}

// UnsafeNotificationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NotificationServer will
// result in compilation errors.
type UnsafeNotificationServer interface {
	mustEmbedUnimplementedNotificationServer()
}

func RegisterNotificationServer(s grpc.ServiceRegistrar, srv NotificationServer) {
	s.RegisterService(&Notification_ServiceDesc, srv)
}

func _Notification_NotifyUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NotifyUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServer).NotifyUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/notification.Notification/NotifyUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServer).NotifyUser(ctx, req.(*NotifyUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notification_BindDeviceToUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BindDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServer).BindDeviceToUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/notification.Notification/BindDeviceToUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServer).BindDeviceToUser(ctx, req.(*BindDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notification_UpdateOldDeviceOnUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateOldDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServer).UpdateOldDeviceOnUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/notification.Notification/UpdateOldDeviceOnUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServer).UpdateOldDeviceOnUser(ctx, req.(*UpdateOldDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notification_UnbindDeviceFromUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnbindDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServer).UnbindDeviceFromUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/notification.Notification/UnbindDeviceFromUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServer).UnbindDeviceFromUser(ctx, req.(*UnbindDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Notification_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/notification.Notification/DeleteUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServer).DeleteUser(ctx, req.(*DeleteUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Notification_ServiceDesc is the grpc.ServiceDesc for Notification service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Notification_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "notification.Notification",
	HandlerType: (*NotificationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NotifyUser",
			Handler:    _Notification_NotifyUser_Handler,
		},
		{
			MethodName: "BindDeviceToUser",
			Handler:    _Notification_BindDeviceToUser_Handler,
		},
		{
			MethodName: "UpdateOldDeviceOnUser",
			Handler:    _Notification_UpdateOldDeviceOnUser_Handler,
		},
		{
			MethodName: "UnbindDeviceFromUser",
			Handler:    _Notification_UnbindDeviceFromUser_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _Notification_DeleteUser_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "notification/notification.proto",
}
