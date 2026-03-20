package handler

import (
	"context"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// IPCheckInterceptor создает UnaryInterceptor для gRPC-сервера,
// который проверяет IP-адрес клиента на соответствие доверенной подсети.
//
// Параметры:
//   - trustedSubnet: строка в формате CIDR (например, "192.168.1.0/24")
//
// Возвращает:
//   - grpc.UnaryServerInterceptor: функция-перехватчик для gRPC
func IPCheckInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	// Если подсеть не задана, возвращаем пустой interceptor
	if trustedSubnet == "" {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	// Парсим CIDR один раз при создании interceptor
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		log.Error().Err(err).Str("trusted_subnet", trustedSubnet).Msg("Failed to parse trusted subnet CIDR")
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, status.Error(codes.Internal, "invalid trusted subnet configuration")
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Получаем метаданные из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Error().Msg("failed to get metadata from context")
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		// Извлекаем IP-адрес клиента
		ips := md["x-real-ip"]
		if len(ips) == 0 {
			log.Error().Msg("x-real-ip metadata not provided")
			return nil, status.Error(codes.InvalidArgument, "x-real-ip metadata not provided")
		}
		clientIP := strings.TrimSpace(ips[0])

		// Проверяем, что IP-адрес не пустой
		if clientIP == "" {
			log.Error().Msg("x-real-ip is empty")
			return nil, status.Error(codes.InvalidArgument, "x-real-ip is empty")
		}

		// Парсим IP-адрес клиента
		parsedIP := net.ParseIP(clientIP)
		if parsedIP == nil {
			log.Error().Str("client_ip", clientIP).Msg("invalid client IP address")
			return nil, status.Error(codes.InvalidArgument, "invalid client IP address")
		}

		// Проверяем принадлежность IP-адреса к доверенной подсети
		if !subnet.Contains(parsedIP) {
			log.Warn().Str("client_ip", clientIP).Str("trusted_subnet", trustedSubnet).Msg("client IP not in trusted subnet")
			return nil, status.Error(codes.PermissionDenied, "client IP not in trusted subnet")
		}

		// IP-адрес прошел проверку, продолжаем обработку
		return handler(ctx, req)
	}
}
