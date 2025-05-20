package tarantool

import (
	"context"
	"fmt"
	"time"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/VK-doodle-jump/internal/errors"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"
)

type UsersStorage struct {
	conn *tarantool.Connection
}

func NewUsersStorage(ctx context.Context, conn *tarantool.Connection) (*UsersStorage, error) {
	storage := &UsersStorage{
		conn: conn,
	}

	return storage, nil
}

func (s *UsersStorage) UpdateUserRating(ctx context.Context, vkid int, newScore int) error {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("user_score").
			Args([]interface{}{map[string]interface{}{
				"vkid": vkid,
			}}).
			Context(ctx),
	).GetResponse()
	if err != nil {
		return fmt.Errorf("(tarantool.UpdateUserRating) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.UpdateUserRating) %w: %v", customErrors.ErrInternal, err)
	}

	var prevScore []int
	err = resp.DecodeTyped(&prevScore)
	if err != nil {
		return fmt.Errorf("(tarantool.UpdateUserRating) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	if prevScore[0] < newScore {
		_, err = s.conn.Do(
			tarantool.NewUpdateRequest("users").
				Index("primary").
				Key([]interface{}{vkid}).
				Operations(tarantool.NewOperations().
					Assign(5, newScore).
					Assign(6, datetime)).
				Context(ctx),
		).Get()

		if err != nil {
			return fmt.Errorf("(tarantool.UpdateUserRating) %w: %v", customErrors.ErrTarantoolExec, err)
		}
	}

	return nil
}

func (s *UsersStorage) GetTopUsers(ctx context.Context, count int) ([]domain.UserRating, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("users_top").
			Args([]interface{}{map[string]interface{}{
				"limit": count,
			}}).
			Context(ctx),
	).GetResponse()
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTopUsers) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data [][]domain.UserRating
	err = resp.DecodeTyped(&data)
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTopUsers) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *UsersStorage) GetNearbyUsers(ctx context.Context, vkid, count int) ([]domain.UserRating, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("users_nearby").
			Args([]interface{}{map[string]interface{}{
				"vkid":  vkid,
				"limit": count,
			}}).
			Context(ctx),
	).GetResponse()
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetNearbyUsers) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data [][]domain.UserRating
	err = resp.DecodeTyped(&data)
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetNearbyUsers) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *UsersStorage) UserScore(ctx context.Context, vkid int) (int, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("user_score").
			Args([]interface{}{map[string]interface{}{
				"vkid": vkid,
			}}).
			Context(ctx),
	).GetResponse()
	if err != nil {
		return 0, fmt.Errorf("(tarantool.UserScore) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var score []int
	err = resp.DecodeTyped(&score)
	if err != nil {
		return 0, fmt.Errorf("(tarantool.UserScore) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return score[0], nil
}
