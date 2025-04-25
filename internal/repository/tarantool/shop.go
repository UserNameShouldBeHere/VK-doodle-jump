package tarantool

import (
	"context"
	"fmt"
	"time"

	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/VK-doodle-jump/internal/errors"
)

type AdminShopStorage struct {
	conn *tarantool.Connection
}

func NewAdminShopStorage(conn *tarantool.Connection) (*AdminShopStorage, error) {
	storage := &AdminShopStorage{
		conn: conn,
	}

	return storage, nil
}

type PromocodeAdminDataT struct {
	Id             int
	Name           string
	Company        string
	LogoLink       string
	Description    string
	Price          int
	Count          int
	Code           string
	ActivationLink string
	ActiveTo       datetime.Datetime
}

func (s *AdminShopStorage) GetPromocodes(ctx context.Context) ([]domain.PromocodeAdminData, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("promocodes_for_admin").
			Context(ctx),
	).GetResponse()
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetPromocodes) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data [][]PromocodeAdminDataT
	err = resp.DecodeTyped(&data)
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetPromocodes) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	promocodes := make([]domain.PromocodeAdminData, len(data[0]))
	for i := range data[0] {
		promocodes[i] = domain.PromocodeAdminData{
			Id:             data[0][i].Id,
			Name:           data[0][i].Name,
			Company:        data[0][i].Company,
			LogoLink:       data[0][i].LogoLink,
			Description:    data[0][i].Description,
			Price:          data[0][i].Price,
			Count:          data[0][i].Count,
			Code:           data[0][i].Code,
			ActivationLink: data[0][i].ActivationLink,
			ActiveTo:       data[0][i].ActiveTo.ToTime(),
		}
	}

	return promocodes, nil
}

func (s *AdminShopStorage) AddPromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error {
	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	activeTo, err := datetime.MakeDatetime(newPromocode.ActiveTo)
	if err != nil {
		return fmt.Errorf("(tarantool.AddPromocode) %w: %v", customErrors.ErrInternal, err)
	}

	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.AddPromocode) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewInsertRequest("promocodes").
			Tuple([]interface{}{
				nil,
				newPromocode.Name,
				newPromocode.Company,
				newPromocode.LogoLink,
				newPromocode.Description,
				newPromocode.Price,
				newPromocode.Count,
				newPromocode.Code,
				newPromocode.ActivationLink,
				activeTo,
				datetime,
			}).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.AddPromocode) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) UpdatePromocode(ctx context.Context, newPromocode domain.PromocodeAdminData) error {
	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	activeTo, err := datetime.MakeDatetime(newPromocode.ActiveTo)
	if err != nil {
		return fmt.Errorf("(tarantool.UpdatePromocode) %w: %v", customErrors.ErrInternal, err)
	}

	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.UpdatePromocode) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewUpdateRequest("promocodes").
			Index("primary").
			Key([]interface{}{newPromocode.Id}).
			Operations(tarantool.NewOperations().
				Assign(1, newPromocode.Name).
				Assign(2, newPromocode.Company).
				Assign(3, newPromocode.LogoLink).
				Assign(4, newPromocode.Description).
				Assign(5, newPromocode.Price).
				Assign(6, newPromocode.Count).
				Assign(7, newPromocode.Code).
				Assign(8, newPromocode.ActivationLink).
				Assign(9, activeTo).
				Assign(10, datetime),
			).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.UpdatePromocode) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) DeletePromocode(ctx context.Context, id int) error {
	_, err := s.conn.Do(
		tarantool.NewDeleteRequest("promocodes").
			Index("primary").
			Key([]interface{}{id}).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.DeletePromocode) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) GetProducts(ctx context.Context) ([]domain.ProductAdminData, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("products_for_admin").
			Context(ctx),
	).GetResponse()
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetProducts) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data [][]domain.ProductAdminData
	err = resp.DecodeTyped(&data)
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetProducts) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *AdminShopStorage) AddProduct(ctx context.Context, newProduct domain.ProductAdminData) error {
	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.AddProduct) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewInsertRequest("products").
			Tuple([]interface{}{
				nil,
				newProduct.Name,
				newProduct.PhotoLink,
				newProduct.Description,
				newProduct.Price,
				newProduct.Count,
				newProduct.ActivationLink,
				datetime,
			}).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.AddProduct) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) UpdateProduct(ctx context.Context, newProduct domain.ProductAdminData) error {
	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.UpdateProduct) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewUpdateRequest("products").
			Index("primary").
			Key([]interface{}{newProduct.Id}).
			Operations(tarantool.NewOperations().
				Assign(1, newProduct.Name).
				Assign(2, newProduct.PhotoLink).
				Assign(3, newProduct.Description).
				Assign(4, newProduct.Price).
				Assign(5, newProduct.Count).
				Assign(6, newProduct.ActivationLink).
				Assign(7, datetime),
			).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.UpdateProduct) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) DeleteProduct(ctx context.Context, id int) error {
	_, err := s.conn.Do(
		tarantool.NewDeleteRequest("products").
			Index("primary").
			Key([]interface{}{id}).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.DeleteProduct) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) GetTasks(ctx context.Context) ([]domain.TaskAdminData, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("tasks_for_admin").
			Context(ctx),
	).GetResponse()
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTasks) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data [][]domain.TaskAdminData
	err = resp.DecodeTyped(&data)
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTasks) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *AdminShopStorage) AddTask(ctx context.Context, newTask domain.TaskAdminData) error {
	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.AddTask) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewInsertRequest("tasks").
			Tuple([]interface{}{
				nil,
				newTask.Name,
				newTask.Description,
				newTask.Reward,
				newTask.Token,
				datetime,
			}).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.AddTask) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) UpdateTask(ctx context.Context, newTask domain.TaskAdminData) error {
	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.UpdateProduct) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewUpdateRequest("tasks").
			Index("primary").
			Key([]interface{}{newTask.Id}).
			Operations(tarantool.NewOperations().
				Assign(1, newTask.Name).
				Assign(2, newTask.Description).
				Assign(3, newTask.Reward).
				Assign(5, datetime),
			).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.UpdateTask) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AdminShopStorage) DeleteTask(ctx context.Context, id int) error {
	_, err := s.conn.Do(
		tarantool.NewDeleteRequest("tasks").
			Index("primary").
			Key([]interface{}{id}).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.DeleteTask) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}
