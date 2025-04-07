package tarantool

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"
)

type UsersStorage struct {
	conn *tarantool.Connection
}

func NewUsersStorage(conn *tarantool.Connection) (*UsersStorage, error) {
	return &UsersStorage{
		conn: conn,
	}, nil
}

func (s *UsersStorage) UpdateUserRating(ctx context.Context, uuid string, newScore int) error {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("user_score").
			Args([]interface{}{map[string]interface{}{
				"name": uuid,
			}}),
	).GetResponse()
	if err != nil {
		return fmt.Errorf("(tarantool.GetTopUsers): %w", err)
	}

	tm := time.Now()

	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		log.Fatal(err)
	}

	var prevScore []int
	err = resp.DecodeTyped(&prevScore)
	if err != nil {
		_, err = s.conn.Do(
			tarantool.NewInsertRequest("users").
				Tuple([]interface{}{uuid, 1, newScore, datetime}),
		).Get()

		if err != nil {
			return fmt.Errorf("(tarantool.GetTopUsers): %w", err)
		}

		return nil
	}

	if prevScore[0] < newScore {
		data, err := s.conn.Do(
			tarantool.NewUpdateRequest("users").
				Index("name").
				Key([]interface{}{uuid}).
				Operations(tarantool.NewOperations().
					Assign(2, newScore).
					Assign(3, datetime)),
		).Get()

		if err != nil {
			return fmt.Errorf("(tarantool.GetTopUsers): %w", err)
		}

		fmt.Println(data)
	}

	return nil
}

func (s *UsersStorage) GetTopUsers(ctx context.Context, count int) ([]domain.LeagueTopUsers, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("league_users_top").
			Args([]interface{}{map[string]interface{}{
				"limit": count,
			}}),
	).GetResponse()
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTopUsers): %w", err)
	}

	var data [][]domain.LeagueTopUsers
	err = resp.DecodeTyped(&data)
	if err != nil {
		return nil, fmt.Errorf("(tarantool.GetTopUsers): %w", err)
	}

	return data[0], nil
}
