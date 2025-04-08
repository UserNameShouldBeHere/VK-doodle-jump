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

func NewUsersStorage(ctx context.Context, conn *tarantool.Connection) (*UsersStorage, error) {
	storage := &UsersStorage{
		conn: conn,
	}

	go func() {
		<-time.After(time.Second * 10)

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
			leagueUsers[league.Id] = append(leagueUsers[league.Id], user.Name)
		}
	}

	for _, league := range settings[0] {
		usersToLUp := make([]string, 0, league.UpCnt)
		if len(leagueUsers[league.Id]) < league.UpCnt {
			for _, user := range leagueUsers[league.Id] {
				usersToLUp = append(usersToLUp, user)
			}
		} else {
			for _, user := range leagueUsers[league.Id][0:league.UpCnt] {
				usersToLUp = append(usersToLUp, user)
			}
		}

		_, err = s.conn.Do(
			tarantool.NewCallRequest("league_change").
				Args([]interface{}{map[string]interface{}{
					"users": usersToLUp,
					"up":    true,
				}}),
		).GetResponse()
		if err != nil {
			return fmt.Errorf("(tarantool.updateLeagues): %w", err)
		}

		usersToLDown := make([]string, 0, league.UpCnt)
		if league.StayCnt != 0 && len(leagueUsers[league.Id])-league.UpCnt > league.StayCnt {
			for _, user := range leagueUsers[league.Id][league.UpCnt+league.StayCnt:] {
				usersToLDown = append(usersToLDown, user)
			}
			_, err = s.conn.Do(
				tarantool.NewCallRequest("league_change").
					Args([]interface{}{map[string]interface{}{
						"users": usersToLDown,
						"up":    false,
					}}),
			).GetResponse()
			if err != nil {
				return fmt.Errorf("(tarantool.updateLeagues): %w", err)
			}
		}
	}

	return nil
}
