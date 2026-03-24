package handler

import (
	"context"
	"inventory-service/internal/service"
	pb "vv-ecommerce/pkg/proto/inventory"
)

type GRPCHandler struct {
	pb.UnimplementedInventoryServiceServer
	svc *service.InventoryService
}

func NewGRPCHandler(svc *service.InventoryService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) DecreaseStock(ctx context.Context, req *pb.DecreaseStockRequest) (*pb.DecreaseStockResponse, error) {
	// gRPC 自动生成的字段是 int64，但我们的 Service 目前用的是 int
	// 在生产环境中，建议全部统一为 int64
	err := h.svc.DecreaseInventory(ctx, req.RequestId, req.Sku, req.OrderId, req.TraceId, int(req.Quantity))
	if err != nil {
		return &pb.DecreaseStockResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DecreaseStockResponse{
		Success: true,
		Message: "success",
	}, nil
}
