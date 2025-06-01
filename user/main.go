package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"shop/user/handler"
	proto "shop/user/proto"
	"syscall"
)

func main() {
	IP := flag.String("ip", "127.0.0.1", "server ip")
	Port := flag.Int("port", 9999, "server port")
	flag.Parse()

	log.Printf("server starting at %s:%d", *IP, *Port)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterUserServiceServer(grpcServer, &handler.UserServer{})

	// 启动服务（并发）
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
}
