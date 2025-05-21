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
