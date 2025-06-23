package tarantool

import (
	"context"
	"fmt"

	"github.com/tarantool/go-tarantool/v2"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/VK-doodle-jump/internal/errors"
)

type ShopStorage struct {
	conn *tarantool.Connection
}

func NewShopStorage(conn *tarantool.Connection) (*ShopStorage, error) {
	storage := &ShopStorage{
		conn: conn,
	}

	return storage, nil
}

func (s *ShopStorage) GetTasks(ctx context.Context, vkid int) ([]domain.TaskData, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("tasks").
			Args([]interface{}{map[string]interface{}{
				"vkid": vkid,
			}}).
			Context(ctx),
	).GetResponse()
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTasks) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data [][]domain.TaskData
	err = resp.DecodeTyped(&data)
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTasks) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *ShopStorage) GetCurrentGiftaway(ctx context.Context) (domain.Giftaway, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("current_giftaway").
			Context(ctx),
	).GetResponse()
	if err != nil {
		return domain.Giftaway{}, fmt.Errorf("(tarantool.GetCurrentGiftaway) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data [][]domain.GiftawayT
	err = resp.DecodeTyped(&data)
	if err != nil {
		return domain.Giftaway{}, fmt.Errorf("(tarantool.GetCurrentGiftaway) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	res := domain.Giftaway{}
	res.Info.Id = data[0][0].Info.Id
	res.Info.Description = data[0][0].Info.Description
	res.Info.Details = data[0][0].Info.Details
	res.Info.From = data[0][0].Info.From.ToTime()
	res.Info.To = data[0][0].Info.To.ToTime()
	res.Gifts = data[0][0].Gifts

	return res, nil
}
