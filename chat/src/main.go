package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/chat/src/config"
	"example.com/chat/src/database"
	"example.com/chat/src/internal/app"
	"example.com/chat/src/internal/client"
	"example.com/chat/src/internal/controller"
	"example.com/chat/src/internal/repository"
	"example.com/chat/src/internal/server"
	"example.com/chat/src/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	database.Init(cfg)
	defer database.Close()

	chatRepository := repository.NewChatRepository(database.DB)
	directChatRepository := repository.NewDirectChatRepository(database.DB)
	chatUserRepository := repository.NewChatUserRepository(database.DB)
	roleRepository := repository.NewRoleRepository(database.DB)
	rolePermissionRepository := repository.NewRolePermissionRepository(database.DB)
	chatRoleRepository := repository.NewChatRoleRepository(database.DB)
	authClient := client.NewAuthClient(cfg)
	usersClient := client.NewUsersClient(cfg)
	chatService := service.NewChatService(
		chatRepository,
		directChatRepository,
		chatUserRepository,
		roleRepository,
		rolePermissionRepository,
		chatRoleRepository,
		usersClient,
	)
	chatController := controller.NewChatController(chatService, authClient)
	grpcServer := server.NewGrpcServer(chatService)
	httpServer := server.NewHttpServer(chatController)
	application := app.New(httpServer, grpcServer, cfg)
	go application.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
