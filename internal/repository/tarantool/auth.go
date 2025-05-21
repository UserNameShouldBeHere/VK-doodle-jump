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

type AuthStorage struct {
	conn *tarantool.Connection
}

func NewAuthStorage(conn *tarantool.Connection) (*AuthStorage, error) {
	storage := &AuthStorage{
		conn: conn,
	}

	return storage, nil
}

func (s *AuthStorage) SignIn(ctx context.Context, signInData domain.SignInData) (bool, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("has_user").
			Args([]interface{}{map[string]interface{}{
				"vkid": signInData.User.VkId,
			}}),
	).GetResponse()
	if err != nil {
		return false, fmt.Errorf("(tarantool.SignIn) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data []bool
	err = resp.DecodeTyped(&data)
	if err != nil {
		return false, fmt.Errorf("(tarantool.SignIn) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	isFirstTime := false
	if data[0] {
		err = s.updateSession(ctx, signInData)
		if err != nil {
			return false, fmt.Errorf("(tarantool.SignIn) %w", err)
		}
	} else {
		err = s.createUser(ctx, signInData)
		if err != nil {
			return false, fmt.Errorf("(tarantool.SignIn) %w", err)
		}
		isFirstTime = true
	}

	return isFirstTime, nil
}

func (s *AuthStorage) Check(ctx context.Context, vkid int, accessToken string) (bool, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("check_user_access").
			Args([]interface{}{map[string]interface{}{
				"vkid":  vkid,
				"token": accessToken,
			}}),
	).GetResponse()
	if err != nil {
		return false, fmt.Errorf("(tarantool.Check) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data []bool
	err = resp.DecodeTyped(&data)
	if err != nil {
		return false, fmt.Errorf("(tarantool.Check) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *AuthStorage) Logout(ctx context.Context, vkid int) error {
	tm := time.Now()
	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.Logout) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewUpdateRequest("users").
			Index("primary").
			Key([]interface{}{vkid}).
			Operations(tarantool.NewOperations().
				Assign(4, datetime),
			).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.Logout) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AuthStorage) GetUserData(ctx context.Context, vkid int) (domain.UserHeader, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("user_header").
			Args([]interface{}{map[string]interface{}{
				"vkid": vkid,
			}}).
			Context(ctx),
	).GetResponse()
	if err != nil {
		return domain.UserHeader{}, fmt.Errorf("(tarantool.GetUserData) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data []domain.UserHeader
	err = resp.DecodeTyped(&data)
	if err != nil {
		return domain.UserHeader{}, fmt.Errorf("(tarantool.GetUserData) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *AuthStorage) IsAdmin(ctx context.Context, vkid int) (bool, error) {
	resp, err := s.conn.Do(
		tarantool.NewCallRequest("is_admin").
			Args([]interface{}{map[string]interface{}{
				"vkid": vkid,
			}}).
			Context(ctx),
	).GetResponse()
	if err != nil {
		return false, fmt.Errorf("(tarantool.IsAdmin) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	var data []bool
	err = resp.DecodeTyped(&data)
	if err != nil {
		return false, fmt.Errorf("(tarantool.IsAdmin) %w: %v", customErrors.ErrTarantoolDecode, err)
	}

	return data[0], nil
}

func (s *AuthStorage) createUser(ctx context.Context, signInData domain.SignInData) error {
	tm := time.Now().Add(time.Hour)
	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	accessExpiration, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.createUser) %w: %v", customErrors.ErrInternal, err)
	}

	tm = time.Now()
	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	datetime, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.createUser) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewInsertRequest("users").
			Tuple([]interface{}{
				signInData.User.VkId,
				signInData.User.Name,
				signInData.User.Avatar,
				signInData.AccessToken,
				accessExpiration,
				0,
				datetime,
			}).
			Context(ctx),
	).Get()

	_, err = s.conn.Do(
		tarantool.NewInsertRequest("game").
			Tuple([]interface{}{
				signInData.User.VkId,
				0,
			}).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.createUser) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}

func (s *AuthStorage) updateSession(ctx context.Context, signInData domain.SignInData) error {
	tm := time.Now().Add(time.Hour)
	tm = tm.In(time.FixedZone(datetime.NoTimezone, 0))
	accessExpiration, err := datetime.MakeDatetime(tm)
	if err != nil {
		return fmt.Errorf("(tarantool.updateSession) %w: %v", customErrors.ErrInternal, err)
	}

	_, err = s.conn.Do(
		tarantool.NewUpdateRequest("users").
			Index("primary").
			Key([]interface{}{signInData.User.VkId}).
			Operations(tarantool.NewOperations().
				Assign(0, signInData.User.VkId).
				Assign(1, signInData.User.Name).
				Assign(2, signInData.User.Avatar).
				Assign(3, signInData.AccessToken).
				Assign(4, accessExpiration),
			).
			Context(ctx),
	).Get()

	if err != nil {
		return fmt.Errorf("(tarantool.updateSession) %w: %v", customErrors.ErrTarantoolExec, err)
	}

	return nil
}
