// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: proto.proto

package pkg

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/timestamp"
	math "math"
)

import (
	context "context"
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for ReporterService service

type ReporterService interface {
	CreateFile(ctx context.Context, in *CreateFileRequest, opts ...client.CallOption) (*CreateFileResponse, error)
	UpdateFile(ctx context.Context, in *UpdateFileRequest, opts ...client.CallOption) (*ResponseError, error)
	GetFile(ctx context.Context, in *GetFileRequest, opts ...client.CallOption) (*GetFileResponse, error)
}

type reporterService struct {
	c    client.Client
	name string
}

func NewReporterService(name string, c client.Client) ReporterService {
	if c == nil {
		c = client.NewClient()
	}
	if len(name) == 0 {
		name = "pkg"
	}
	return &reporterService{
		c:    c,
		name: name,
	}
}

func (c *reporterService) CreateFile(ctx context.Context, in *CreateFileRequest, opts ...client.CallOption) (*CreateFileResponse, error) {
	req := c.c.NewRequest(c.name, "ReporterService.CreateFile", in)
	out := new(CreateFileResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reporterService) UpdateFile(ctx context.Context, in *UpdateFileRequest, opts ...client.CallOption) (*ResponseError, error) {
	req := c.c.NewRequest(c.name, "ReporterService.UpdateFile", in)
	out := new(ResponseError)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reporterService) GetFile(ctx context.Context, in *GetFileRequest, opts ...client.CallOption) (*GetFileResponse, error) {
	req := c.c.NewRequest(c.name, "ReporterService.GetFile", in)
	out := new(GetFileResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ReporterService service

type ReporterServiceHandler interface {
	CreateFile(context.Context, *CreateFileRequest, *CreateFileResponse) error
	UpdateFile(context.Context, *UpdateFileRequest, *ResponseError) error
	GetFile(context.Context, *GetFileRequest, *GetFileResponse) error
}

func RegisterReporterServiceHandler(s server.Server, hdlr ReporterServiceHandler, opts ...server.HandlerOption) error {
	type reporterService interface {
		CreateFile(ctx context.Context, in *CreateFileRequest, out *CreateFileResponse) error
		UpdateFile(ctx context.Context, in *UpdateFileRequest, out *ResponseError) error
		GetFile(ctx context.Context, in *GetFileRequest, out *GetFileResponse) error
	}
	type ReporterService struct {
		reporterService
	}
	h := &reporterServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&ReporterService{h}, opts...))
}

type reporterServiceHandler struct {
	ReporterServiceHandler
}

func (h *reporterServiceHandler) CreateFile(ctx context.Context, in *CreateFileRequest, out *CreateFileResponse) error {
	return h.ReporterServiceHandler.CreateFile(ctx, in, out)
}

func (h *reporterServiceHandler) UpdateFile(ctx context.Context, in *UpdateFileRequest, out *ResponseError) error {
	return h.ReporterServiceHandler.UpdateFile(ctx, in, out)
}

func (h *reporterServiceHandler) GetFile(ctx context.Context, in *GetFileRequest, out *GetFileResponse) error {
	return h.ReporterServiceHandler.GetFile(ctx, in, out)
}
