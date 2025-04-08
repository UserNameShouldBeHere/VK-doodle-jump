package tarantool

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
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

	go func() {
		// <-time.After(time.Second * 10)

		for {
			select {
			case <-time.After(time.Second * 10):
				err := storage.updateLeagues()
				if err != nil {
					fmt.Println(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return storage, nil
}

func (s *UsersStorage) UpdateUserRating(ctx context.Context, uuid string, newScore int) error {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("user_score").
			Args([]interface{}{map[string]interface{}{
				"name": uuid,
			}}),
	).GetResponse()
	if err != nil {
		return fmt.Errorf("(tarantool.UpdateUserRating): %w", err)
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
				Tuple([]interface{}{uuid, 0, newScore, datetime}),
		).Get()

		if err != nil {
			return fmt.Errorf("(tarantool.UpdateUserRating): %w", err)
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
			return fmt.Errorf("(tarantool.UpdateUserRating): %w", err)
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

func (s *UsersStorage) updateLeagues() error {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("leagues_settings").
			Args([]interface{}{}),
	).GetResponse()
	if err != nil {
		return fmt.Errorf("(tarantool.updateLeagues): %w", err)
	}

	var settings [][]struct {
		Id      int
		UpCnt   int
		StayCnt int
	}
	err = resp.DecodeTyped(&settings)
	if err != nil {
		return fmt.Errorf("(tarantool.updateLeagues): %w", err)
	}

	leagueUsers := make([][]string, len(settings[0]))
	for _, league := range settings[0] {
		resp, err = s.conn.Do(
			tarantool.NewCallRequest("league_users_pos").
				Args([]interface{}{map[string]interface{}{
					"league": league.Id,
				}}),
		).GetResponse()
		if err != nil {
			return fmt.Errorf("(tarantool.updateLeagues): %w", err)
		}

		var users [][]struct {
			Name string
		}
		err = resp.DecodeTyped(&users)
		if err != nil {
			return fmt.Errorf("(tarantool.updateLeagues): %w", err)
		}

		for _, user := range users[0] {
			leagueUsers[len(settings[0])-league.Id-1] = append(leagueUsers[len(settings[0])-league.Id-1], user.Name)
		}
	}

	fmt.Println(leagueUsers)

	upUsers := make([]string, 0)
	downUsers := make([]string, 0)
	for i, league := range leagueUsers[1:] {
		cnt := math.Min(float64(len(league)), float64(settings[0][i+1].UpCnt))
		users := league[0:int(cnt)]
		leagueUsers[i] = append(users, leagueUsers[i]...)
		leagueUsers[i+1] = leagueUsers[i+1][int(cnt):]

		fmt.Printf("Up(%d): %v\n", 3-i, users)

		upUsers = append(upUsers, users...)

		if settings[0][i].StayCnt < len(leagueUsers[i]) {
			cnt := len(leagueUsers[i]) - settings[0][i].StayCnt
			if len(leagueUsers[i])-cnt < len(users) {
				users = leagueUsers[i][len(users):]
			} else {
				users = leagueUsers[i][len(leagueUsers[i])-cnt:]
			}
			leagueUsers[i+1] = append(users, leagueUsers[i+1]...)

			fmt.Printf("Down(%d): %v\n", 3-i, users)

			downUsers = append(downUsers, users...)
		}
	}

	_, err = s.conn.Do(
		tarantool.NewCallRequest("league_change").
			Args([]interface{}{map[string]interface{}{
				"users": upUsers,
				"up":    true,
			}}),
	).GetResponse()
	if err != nil {
		return fmt.Errorf("(tarantool.updateLeagues): %w", err)
	}

	_, err = s.conn.Do(
		tarantool.NewCallRequest("league_change").
			Args([]interface{}{map[string]interface{}{
				"users": downUsers,
				"up":    false,
			}}),
	).GetResponse()
	if err != nil {
		return fmt.Errorf("(tarantool.updateLeagues): %w", err)
	}

	return nil
}
