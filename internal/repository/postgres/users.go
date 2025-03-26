package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/UserNameShouldBeHere/VK-doodle-jump/internal/domain"
)

type UsersStorage struct {
	pgConn PgConn
}

func NewUsersStorage(pgConn PgConn) (*UsersStorage, error) {
	return &UsersStorage{
		pgConn: pgConn,
	}, nil
}

func (s *UsersStorage) UpdateUserRating(ctx context.Context, uuid string, newScore int) error {
	tx, err := s.pgConn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("(postgres.UpdateUserRating): %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			fmt.Printf("(postgres.UpdateUserRating): %v", err)
		}
	}()

	var prevScore int
	err = s.pgConn.QueryRow(ctx, `
		select max_score
		from rating
		where name = $1;
	`, uuid).Scan(&prevScore)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.pgConn.Exec(ctx, `
				insert into rating (name, max_score) values ($1, $2);
			`, uuid, newScore)
			if err != nil {
				return fmt.Errorf("(postgres.UpdateUserRating): %w", err)
			}

			return nil
		}

		return fmt.Errorf("(postgres.UpdateUserRating): %w", err)
	}

	if prevScore < newScore {
		_, err = s.pgConn.Exec(ctx, `
			update rating set max_score = $1 where name = $2;
		`, newScore, uuid)
		if err != nil {
			return fmt.Errorf("(postgres.UpdateUserRating): %w", err)
		}
		_, err = s.pgConn.Exec(ctx, `
			update rating set last_update = now() where name = $1;
		`, uuid)
		if err != nil {
			return fmt.Errorf("(postgres.UpdateUserRating): %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("(postgres.UpdateUserRating): %w", err)
	}

	return nil
}

func (s *UsersStorage) GetTopUsers(ctx context.Context, count int) ([]domain.UserRating, error) {
	users := make([]domain.UserRating, 0)

	rows, err := s.pgConn.Query(ctx, `
		select name, max_score
		from rating
		order by max_score desc, last_update asc
		limit $1;
	`, count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users, nil
		}

		return nil, fmt.Errorf("(postgres.GetTopUsers): %w", err)
	}

	for rows.Next() {
		var user domain.UserRating

		err = rows.Scan(&user.Name, &user.Score)
		if err != nil {
			return nil, fmt.Errorf("(postgres.GetTopUsers): %w", err)
		}

		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("(postgres.GetTopUsers): %w", err)
	}

	return users, nil
}
