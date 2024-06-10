package server

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"

	"example.com/media-handler/src/gen/go/media"
	"example.com/media-handler/src/internal/client"
	"example.com/media-handler/src/internal/controller"
	"example.com/media-handler/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HttpServer struct {
	mediaHandlerController *controller.MediaHandlerController
}

func NewHttpServer(mediaHandlerController *controller.MediaHandlerController) *HttpServer {
	return &HttpServer{mediaHandlerController: mediaHandlerController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("POST /upload", h.mediaHandlerController.UploadMediaHandler)
	http.HandleFunc("GET /uploads/{id}", h.mediaHandlerController.GetMediaHandler)
	http.HandleFunc("DELETE /delete/{id}", h.mediaHandlerController.DeleteMediaHandler)
}

type GRPCServer struct {
	gRPCServer *grpc.Server
	media.UnimplementedMediaHandlerServer
	mediaHandlerService *service.MediaHandlerService
	authClient          *client.AuthGRPCClient
}

func NewGRPCServer(
	mediaHandlerService *service.MediaHandlerService,
	authClient *client.AuthGRPCClient,
) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer:          gRPCServer,
		mediaHandlerService: mediaHandlerService,
		authClient:          authClient,
	}
	media.RegisterMediaHandlerServer(g.gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) StoreImage(stream media.MediaHandler_StoreImageServer) error {
	var fileName string
	file, err := os.CreateTemp("", "tmp_")
	if err != nil {
		slog.Error(err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	defer os.Remove(file.Name())
	defer file.Close()

	ctx := stream.Context()
	_, err = s.authClient.PerformAuthorize(ctx, nil)
	if err != nil {
		slog.Error(err.Error())
		return status.Error(codes.PermissionDenied, err.Error())
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error(err.Error())
			return status.Error(codes.InvalidArgument, err.Error())
		}

		file.Write(req.Data)

		fileName = req.FileName
	}

	if fileName == "" {
		return status.Error(codes.InvalidArgument, "FileName is empty")
	}

	fileId, err := s.mediaHandlerService.UpdateAvatar(file, fileName)
	if fileId == uuid.Nil || err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	resp := &media.ImageResponse{
		FileId: fileId.String(),
	}
	return stream.SendAndClose(resp)
}
