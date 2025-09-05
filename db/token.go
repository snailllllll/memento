package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserToken 用户token结构体
type UserToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // token ID
	Name      string             `bson:"name" json:"name"`                  // 用户名
	Channel   string             `bson:"channel" json:"channel"`            // 渠道 (qq/phone)
	Token     string             `bson:"token" json:"token"`                // token字符串
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`      // 过期时间
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`      // 创建时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`      // 更新时间
}

// TokenService token服务
type TokenService struct {
	collection *mongo.Collection
}

// NewTokenService 创建token服务
func NewTokenService() *TokenService {
	return &TokenService{
		collection: Collection("message_db", "user_tokens"),
	}
}

// GenerateUserToken 生成用户token并存储到MongoDB
func (s *TokenService) GenerateUserToken(ctx context.Context, name, channel string) (*UserToken, error) {
	// 生成随机token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.New("生成token失败")
	}
	token := hex.EncodeToString(tokenBytes)

	// 设置过期时间为30天后
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	userToken := &UserToken{
		Name:      name,
		Channel:   channel,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 插入新token
	result, err := s.collection.InsertOne(ctx, userToken)
	if err != nil {
		return nil, err
	}

	// 获取插入的ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		userToken.ID = oid
	}

	return userToken, nil
}

// ValidateToken 验证token是否有效，如果有效则更新过期时间
func (s *TokenService) ValidateToken(ctx context.Context, token string) (bool, *UserToken, error) {
	var userToken UserToken
	filter := bson.M{"token": token}
	err := s.collection.FindOne(ctx, filter).Decode(&userToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil, nil
		}
		return false, nil, err
	}

	// 检查是否过期
	if time.Now().After(userToken.ExpiresAt) {
		return false, nil, nil
	}

	// 更新过期时间（延长30天）
	newExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	update := bson.M{
		"$set": bson.M{
			"expires_at": newExpiresAt,
			"updated_at": time.Now(),
		},
	}

	_, err = s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, nil, err
	}

	// 更新内存中的过期时间
	userToken.ExpiresAt = newExpiresAt
	return true, &userToken, nil
}

// GetUserByToken 从token获取用户信息
func (s *TokenService) GetUserByToken(ctx context.Context, token string) (*UserToken, error) {
	var userToken UserToken
	filter := bson.M{"token": token}
	err := s.collection.FindOne(ctx, filter).Decode(&userToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("token不存在")
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(userToken.ExpiresAt) {
		return nil, errors.New("token已过期")
	}

	return &userToken, nil
}

// GetUserFromContext 从gin.Context的token字段中反向查找用户名
func (s *TokenService) GetUserFromContext(ctx context.Context, c interface{}) (string, error) {
	// 这里假设c是gin.Context类型，通过反射获取token
	// 由于不能直接导入gin包，这里使用interface{}和类型断言
	if ctxValue, ok := c.(interface {
		Get(string) (interface{}, bool)
	}); ok {
		if tokenValue, exists := ctxValue.Get("token"); exists {
			if token, ok := tokenValue.(string); ok {
				userToken, err := s.GetUserByToken(ctx, token)
				if err != nil {
					return "", err
				}
				return userToken.Name, nil
			}
		}
	}
	return "", errors.New("无法从context获取token")
}

// DeleteToken 删除token
func (s *TokenService) DeleteToken(ctx context.Context, token string) error {
	filter := bson.M{"token": token}
	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("token不存在")
	}
	return nil
}

// CreateIndexes 创建索引
func (s *TokenService) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.M{"token": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"name": 1, "channel": 1},
		},
		{
			Keys: bson.M{"expires_at": 1},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// CleanExpiredTokens 清理过期token
func (s *TokenService) CleanExpiredTokens(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lt": time.Now()}}
	_, err := s.collection.DeleteMany(ctx, filter)
	return err
}
