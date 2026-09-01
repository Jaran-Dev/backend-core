package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
	"github.com/Jaran-Dev/backend-core/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(client *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{client: client}
}

func (s *RedisSessionStore) SaveSession(ctx context.Context, sessionID uuid.UUID, hashedToken string, accID account.AccountID, role auth.Role, expiresAt time.Time, meta auth.SessionMetadata) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return auth.ErrTokenExpired
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID.String())
	tokenIndexKey := fmt.Sprintf("token_index:%s", hashedToken)
	accountSessionsKey := fmt.Sprintf("account_sessions:%s", accID.String())

	now := time.Now().Unix()

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, sessionKey,
		"token_hash", hashedToken,
		"account_id", accID.String(),
		"role", string(role),
		"used", "0",
		"expires_at", expiresAt.Unix(),
		"created_at", now,
		"last_active_at", now,
		"ip_address", meta.IPAddress,
		"user_agent", meta.UserAgent,
		"device_type", meta.DeviceType,
	)
	pipe.Expire(ctx, sessionKey, ttl)

	pipe.Set(ctx, tokenIndexKey, sessionID.String(), ttl)

	pipe.ZAdd(ctx, accountSessionsKey, redis.Z{
		Score:  float64(expiresAt.Unix()),
		Member: sessionID.String(),
	})
	pipe.Expire(ctx, accountSessionsKey, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisSessionStore) GetSession(ctx context.Context, hashedToken string) (uuid.UUID, account.AccountID, auth.Role, bool, time.Time, error) {
	tokenIndexKey := fmt.Sprintf("token_index:%s", hashedToken)
	sessionIDStr, err := s.client.Get(ctx, tokenIndexKey).Result()
	if err == redis.Nil {
		return uuid.Nil, account.AccountID(uuid.Nil), "", false, time.Time{}, auth.ErrTokenNotFound
	}
	if err != nil {
		return uuid.Nil, account.AccountID(uuid.Nil), "", false, time.Time{}, err
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return uuid.Nil, account.AccountID(uuid.Nil), "", false, time.Time{}, auth.ErrTokenNotFound
	}

	sessionKey := fmt.Sprintf("session:%s", sessionIDStr)
	res, err := s.client.HGetAll(ctx, sessionKey).Result()
	if err != nil || len(res) == 0 {
		return uuid.Nil, account.AccountID(uuid.Nil), "", false, time.Time{}, auth.ErrTokenNotFound
	}

	uid, _ := uuid.Parse(res["account_id"])
	role := auth.Role(res["role"])
	used := res["used"] == "1"
	expUnix, _ := strconv.ParseInt(res["expires_at"], 10, 64)
	exp := time.Unix(expUnix, 0)

	return sessionID, account.AccountID(uid), role, used, exp, nil
}

const rotateScript = `
local old_hash = KEYS[1]
local new_hash = ARGV[1]
local now = tonumber(ARGV[2])
local new_exp = tonumber(ARGV[3])
local ip = ARGV[4]
local user_agent = ARGV[5]
local device_type = ARGV[6]

local old_session_id = redis.call('GET', 'token_index:' .. old_hash)
if not old_session_id then return {err='not_found'} end

local session_key = 'session:' .. old_session_id
local account_id = redis.call('HGET', session_key, 'account_id')
local role = redis.call('HGET', session_key, 'role')
local used = redis.call('HGET', session_key, 'used')
local exp = tonumber(redis.call('HGET', session_key, 'expires_at'))

if used == '1' then
  redis.call('DEL', session_key)
  redis.call('DEL', 'token_index:' .. old_hash)
  redis.call('ZREM', 'account_sessions:' .. account_id, old_session_id)
  return {err='reuse'}
end

if exp < now then return {err='expired'} end

local ttl = new_exp - now
if ttl <= 0 then return {err='expired'} end

redis.call('HSET', session_key, 'used', '1')
redis.call('DEL', 'token_index:' .. old_hash)

redis.call('HSET', session_key, 
    'token_hash', new_hash, 
    'expires_at', new_exp, 
    'last_active_at', now,
    'ip_address', ip,
    'user_agent', user_agent,
    'device_type', device_type)

redis.call('SET', 'token_index:' .. new_hash, old_session_id, 'EX', ttl)
redis.call('EXPIRE', session_key, ttl)

redis.call('ZADD', 'account_sessions:' .. account_id, new_exp, old_session_id)

return {old_session_id, account_id, role}
`

func (s *RedisSessionStore) AtomicRotate(ctx context.Context, hashedOldToken, hashedNewToken string, newExpiresAt time.Time, meta auth.SessionMetadata) (uuid.UUID, account.AccountID, auth.Role, error) {
	now := time.Now().Unix()
	newExp := newExpiresAt.Unix()

	res, err := s.client.Eval(ctx, rotateScript, []string{hashedOldToken},
		hashedNewToken, now, newExp, meta.IPAddress, meta.UserAgent, meta.DeviceType,
	).Result()

	if err != nil {
		if err.Error() == "not_found" {
			return uuid.Nil, account.AccountID(uuid.Nil), "", auth.ErrTokenNotFound
		}
		if err.Error() == "reuse" {
			return uuid.Nil, account.AccountID(uuid.Nil), "", auth.ErrRefreshTokenReused
		}
		if err.Error() == "expired" {
			return uuid.Nil, account.AccountID(uuid.Nil), "", auth.ErrTokenExpired
		}
		return uuid.Nil, account.AccountID(uuid.Nil), "", err
	}

	vals := res.([]interface{})
	sessionID, _ := uuid.Parse(vals[0].(string))
	uid, _ := uuid.Parse(vals[1].(string))
	role := auth.Role(vals[2].(string))

	return sessionID, account.AccountID(uid), role, nil
}

func (s *RedisSessionStore) RevokeSessionByID(ctx context.Context, accID account.AccountID, sessionID uuid.UUID) error {
	sessionKey := fmt.Sprintf("session:%s", sessionID.String())
	accountSessionsKey := fmt.Sprintf("account_sessions:%s", accID.String())

	pipe := s.client.TxPipeline()

	tokenHash, err := s.client.HGet(ctx, sessionKey, "token_hash").Result()
	if err == nil {
		pipe.Del(ctx, fmt.Sprintf("token_index:%s", tokenHash))
	}
	pipe.Del(ctx, sessionKey)
	pipe.ZRem(ctx, accountSessionsKey, sessionID.String())

	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisSessionStore) RevokeAllByAccountID(ctx context.Context, accID account.AccountID) error {
	accountSessionsKey := fmt.Sprintf("account_sessions:%s", accID.String())
	members, err := s.client.ZRange(ctx, accountSessionsKey, 0, -1).Result()
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}

	pipe := s.client.TxPipeline()
	for _, sessionID := range members {
		sessionKey := fmt.Sprintf("session:%s", sessionID)
		tokenHash, err := s.client.HGet(ctx, sessionKey, "token_hash").Result()
		if err == nil {
			pipe.Del(ctx, fmt.Sprintf("token_index:%s", tokenHash))
		}
		pipe.Del(ctx, sessionKey)
	}
	pipe.Del(ctx, accountSessionsKey)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisSessionStore) ListSessionsByAccountID(ctx context.Context, accID account.AccountID) ([]*auth.Session, error) {
	accountSessionsKey := fmt.Sprintf("account_sessions:%s", accID.String())
	now := float64(time.Now().Unix())

	sessionIDs, err := s.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     accountSessionsKey,
		Start:   fmt.Sprintf("%v", now),
		Stop:    "+inf",
		ByScore: true,
	}).Result()
	if err != nil {
		return nil, err
	}

	if len(sessionIDs) == 0 {
		return []*auth.Session{}, nil
	}

	pipe := s.client.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(sessionIDs))
	for i, id := range sessionIDs {
		cmds[i] = pipe.HGetAll(ctx, fmt.Sprintf("session:%s", id))
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var sessions []*auth.Session
	for i, cmd := range cmds {
		res, err := cmd.Result()
		if err != nil || len(res) == 0 {
			continue
		}

		sessionID, _ := uuid.Parse(sessionIDs[i])
		uid, _ := uuid.Parse(res["account_id"])
		expUnix, _ := strconv.ParseInt(res["expires_at"], 10, 64)
		lastActiveUnix, _ := strconv.ParseInt(res["last_active_at"], 10, 64)
		createdUnix, _ := strconv.ParseInt(res["created_at"], 10, 64)

		sessions = append(sessions, &auth.Session{
			ID:        sessionID,
			AccountID: account.AccountID(uid),
			Role:      auth.Role(res["role"]),
			ExpiresAt: time.Unix(expUnix, 0),
			Metadata: auth.SessionMetadata{
				IPAddress:  res["ip_address"],
				UserAgent:  res["user_agent"],
				DeviceType: res["device_type"],
			},
			LastActiveAt: time.Unix(lastActiveUnix, 0),
			CreatedAt:    time.Unix(createdUnix, 0),
		})
	}

	return sessions, nil
}

func (s *RedisSessionStore) SavePasswordResetToken(ctx context.Context, accID account.AccountID, hashedToken string, expiresAt time.Time) error {
	key := fmt.Sprintf("pwd_reset:%s", hashedToken)
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("token already expired")
	}
	return s.client.Set(ctx, key, accID.String(), ttl).Err()
}

func (s *RedisSessionStore) GetAccountIDByPasswordResetToken(ctx context.Context, hashedToken string) (account.AccountID, error) {
	key := fmt.Sprintf("pwd_reset:%s", hashedToken)
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return account.AccountID(uuid.Nil), errors.New("invalid or expired reset token")
	}
	if err != nil {
		return account.AccountID(uuid.Nil), err
	}

	uid, _ := uuid.Parse(val)
	return account.AccountID(uid), nil
}

func (s *RedisSessionStore) DeletePasswordResetToken(ctx context.Context, hashedToken string) error {
	key := fmt.Sprintf("pwd_reset:%s", hashedToken)
	return s.client.Del(ctx, key).Err()
}
