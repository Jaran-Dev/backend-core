package cache

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *RedisSessionStore) SaveSession(ctx context.Context, hashedToken string, accID account.AccountID, role auth.Role, familyID string, expiresAt time.Time) error {
	key := fmt.Sprintf("session:%s", hashedToken)
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return auth.ErrTokenExpired
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, "account_id", accID.String(), "role", string(role), "family_id", familyID, "used", "0", "expires_at", expiresAt.Unix())
	pipe.Expire(ctx, key, ttl)
	pipe.SAdd(ctx, fmt.Sprintf("family:%s", familyID), hashedToken)

	pipe.SAdd(ctx, fmt.Sprintf("account_sessions:%s", accID.String()), hashedToken)
	pipe.Expire(ctx, fmt.Sprintf("account_sessions:%s", accID.String()), ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisSessionStore) GetSession(ctx context.Context, hashedToken string) (account.AccountID, auth.Role, string, bool, time.Time, error) {
	key := fmt.Sprintf("session:%s", hashedToken)
	res, err := s.client.HGetAll(ctx, key).Result()
	if err == redis.Nil || len(res) == 0 {
		return account.AccountID(uuid.Nil), "", "", false, time.Time{}, auth.ErrTokenNotFound
	}
	if err != nil {
		return account.AccountID(uuid.Nil), "", "", false, time.Time{}, err
	}

	uid, _ := uuid.Parse(res["account_id"])
	role := auth.Role(res["role"])
	familyID := res["family_id"]
	used := res["used"] == "1"
	expUnix, _ := strconv.ParseInt(res["expires_at"], 10, 64)
	exp := time.Unix(expUnix, 0)

	return account.AccountID(uid), role, familyID, used, exp, nil
}

// AtomicRotate uses a Lua script to prevent TOCTOU races during refresh token rotation.
const rotateScript = `
local old_key = 'session:' .. KEYS[1]
local new_key = 'session:' .. ARGV[1]
local now = tonumber(ARGV[2])
local new_exp = tonumber(ARGV[3])

if redis.call('EXISTS', old_key) == 0 then return {err='not_found'} end

local account_id = redis.call('HGET', old_key, 'account_id')
local role = redis.call('HGET', old_key, 'role')
local family_id = redis.call('HGET', old_key, 'family_id')
local used = redis.call('HGET', old_key, 'used')
local exp = tonumber(redis.call('HGET', old_key, 'expires_at'))

if used == '1' then
  local fam_key = 'family:' .. family_id
  local members = redis.call('SMEMBERS', fam_key)
  for _, m in ipairs(members) do redis.call('DEL', 'session:' .. m) end
  redis.call('DEL', fam_key)
  return {err='reuse'}
end

if exp < now then return {err='expired'} end

redis.call('HSET', old_key, 'used', '1')

redis.call('HSET', new_key, 'account_id', account_id, 'role', role, 'family_id', family_id, 'used', '0', 'expires_at', new_exp)
redis.call('EXPIREAT', new_key, new_exp)
redis.call('SADD', 'family:' .. family_id, ARGV[1])

return {account_id, role, family_id}
`

func (s *RedisSessionStore) AtomicRotate(ctx context.Context, hashedOldToken, hashedNewToken string, newExpiresAt time.Time) (account.AccountID, auth.Role, error) {
	now := time.Now().Unix()
	newExp := newExpiresAt.Unix()

	res, err := s.client.Eval(ctx, rotateScript, []string{hashedOldToken}, hashedNewToken, now, newExp).Result()
	if err != nil {
		if err.Error() == "not_found" {
			return account.AccountID(uuid.Nil), "", auth.ErrTokenNotFound
		}
		if err.Error() == "reuse" {
			return account.AccountID(uuid.Nil), "", auth.ErrRefreshTokenReused
		}
		if err.Error() == "expired" {
			return account.AccountID(uuid.Nil), "", auth.ErrTokenExpired
		}
		return account.AccountID(uuid.Nil), "", err
	}

	vals := res.([]interface{})
	uid, _ := uuid.Parse(vals[0].(string))
	role := auth.Role(vals[1].(string))
	return account.AccountID(uid), role, nil
}

func (s *RedisSessionStore) RevokeFamily(ctx context.Context, familyID string) error {
	famKey := fmt.Sprintf("family:%s", familyID)
	members, err := s.client.SMembers(ctx, famKey).Result()
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}
	keys := make([]string, len(members))
	for i, m := range members {
		keys[i] = fmt.Sprintf("session:%s", m)
	}
	keys = append(keys, famKey)
	return s.client.Del(ctx, keys...).Err()
}

func (s *RedisSessionStore) RevokeAllByAccountID(ctx context.Context, accID account.AccountID) error {
	sessionsKey := fmt.Sprintf("account_sessions:%s", accID.String())
	members, err := s.client.SMembers(ctx, sessionsKey).Result()
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}

	keys := make([]string, len(members))
	for i, m := range members {
		keys[i] = fmt.Sprintf("session:%s", m)
	}
	keys = append(keys, sessionsKey)
	return s.client.Del(ctx, keys...).Err()
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
