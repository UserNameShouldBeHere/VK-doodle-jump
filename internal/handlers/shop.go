package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"go.uber.org/zap"
)

type ShopService interface {
	GetPromocodes(ctx context.Context) ([]domain.PromocodeAdminData, error)
	AddPromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error
	UpdatePromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error
	GetProducts(ctx context.Context) ([]domain.ProductAdminData, error)
	AddProduct(ctx context.Context, newProduct domain.ProductAdminData) error
	UpdateProduct(ctx context.Context, newProduct domain.ProductAdminData) error
}

type ShopHandler struct {
	shopService ShopService
	logger      *zap.SugaredLogger
}

func NewShopHandler(shopService ShopService, logger *zap.SugaredLogger) (*ShopHandler, error) {
	return &ShopHandler{
		shopService: shopService,
		logger:      logger,
	}, nil
}

type PromocodesAdminResponse struct {
	Promocodes []domain.PromocodeAdminData `json:"promocodes"`
}

func (h *ShopHandler) GetPromocodes(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	promocodes, err := h.shopService.GetPromocodes(ctx)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}

		return
	}

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data: PromocodesAdminResponse{
			Promocodes: promocodes,
		},
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

type UpdatePromocodeRequest struct {
	Promocode domain.PromocodeAdminData `json:"promocode"`
}

func (h *ShopHandler) AddPromocode(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var reqData UpdatePromocodeRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	err = h.shopService.AddPromocode(ctx, reqData.Promocode)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}

		return
	}

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   nil,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

func (h *ShopHandler) UpdatePromocode(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var reqData UpdatePromocodeRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	err = h.shopService.UpdatePromocode(ctx, reqData.Promocode)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}

		return
	}

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   nil,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

type ProductsAdminResponse struct {
	Products []domain.ProductAdminData `json:"products"`
}

func (h *ShopHandler) GetProducts(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	products, err := h.shopService.GetProducts(ctx)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}

		return
	}

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data: ProductsAdminResponse{
			Products: products,
		},
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

type UpdateProductRequest struct {
	Product domain.ProductAdminData `json:"product"`
}

func (h *ShopHandler) AddProduct(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var reqData UpdateProductRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	err = h.shopService.AddProduct(ctx, reqData.Product)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}

		return
	}

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   nil,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}

func (h *ShopHandler) UpdateProduct(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var reqData UpdateProductRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
		return
	}

	err = h.shopService.UpdateProduct(ctx, reqData.Product)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: http.StatusBadRequest,
			Data:   nil,
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}

		return
	}

	err = WriteResponse(w, ResponseData{
		Status: http.StatusOK,
		Data:   nil,
	})
	if err != nil {
		h.logger.Errorf("error at writing response: %v", err)
	}
}
