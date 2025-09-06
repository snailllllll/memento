package verification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"memento_backend/db"

	"snail.local/snailllllll/utils/sms"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"snail.local/snailllllll/napcat_go_sdk"
	"snail.local/snailllllll/utils"
)

// VerificationCode 验证码结构体
type VerificationCode struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Channel   string             `bson:"channel" json:"channel"`
	Code      string             `bson:"code" json:"code"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// VerificationCodeService 验证码服务
type VerificationCodeService struct {
	collection  *mongo.Collection
	userService *db.UserService
}

// NewVerificationCodeService 创建验证码服务
func NewVerificationCodeService() *VerificationCodeService {
	return &VerificationCodeService{
		collection:  db.Collection("message_db", "verification_codes"),
		userService: db.NewUserService(),
	}
}

// GenerateVerificationCode 生成验证码
func (s *VerificationCodeService) GenerateVerificationCode(ctx context.Context, name, channel string) (string, error) {
	var lockTimeout time.Duration
	switch channel {
	case "sms":
		lockTimeout = 2 * time.Minute
	case "qq":
		lockTimeout = 5 * time.Second
	default:
		lockTimeout = 2 * time.Minute
	}

	// 创建锁
	lockKey := fmt.Sprintf("verification_lock_%s_%s", name, channel)

	// 使用TryLock尝试获取锁，如果锁已存在则直接返回错误
	err := utils.TryLock(lockKey, lockTimeout)
	if err != nil {
		return "", errors.New("操作过于频繁，请稍后再试")
	}

	// 注意：这里不立即释放锁，锁会在TTL时间后自动过期
	// 如果需要手动释放锁，可以在适当的时候调用 utils.DeleteLock(lockKey)

	// 生成6位验证码
	code := fmt.Sprintf("%06d", utils.GenerateRandomNumber(100000, 999999))

	// 设置过期时间为5分钟
	expiresAt := time.Now().Add(5 * time.Minute)

	// 创建验证码记录
	verificationCode := &VerificationCode{
		Name:      name,
		Channel:   channel,
		Code:      code,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// 插入数据库
	_, err = s.collection.InsertOne(ctx, verificationCode)
	if err != nil {
		// 如果数据库操作失败，可以选择手动释放锁
		utils.DeleteLock(lockKey)
		return "", fmt.Errorf("创建验证码失败: %v", err)
	}

	// 获取用户信息
	user, err := s.userService.GetUserByName(ctx, name)
	if err != nil {
		// 如果获取用户信息失败，可以选择手动释放锁
		utils.DeleteLock(lockKey)
		return "", fmt.Errorf("获取用户信息失败: %v", err)
	}
	message := fmt.Sprintf("【翻旧账】您的验证码是：%s，有效期 15分钟，请勿泄露给他人。", code)

	// 根据channel选择发送方式
	if channel == "qq" {
		// 发送验证码到QQ
		if user.QQ != "" {
			ws, _ := napcat_go_sdk.GetExistWSClient()
			napcat_go_sdk.SingleTextMessage(&message, &user.QQ, ws)
		}
	} else {
		// 发送验证码到手机
		if user.Phone != "" {
			client := sms.NewClient()
			err := client.SendSMS(2, user.Phone, message)
			if err != nil {
				// 记录错误但不中断流程
				fmt.Printf("发送短信失败: %v", err)
			}
		}
	}

	return code, nil
}

// VerifyCode 验证验证码
func (s *VerificationCodeService) VerifyCode(ctx context.Context, name, channel, code string) (bool, error) {
	filter := bson.M{
		"name":       name,
		"channel":    channel,
		"code":       code,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	var verificationCode VerificationCode
	err := s.collection.FindOne(ctx, filter).Decode(&verificationCode)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	// 验证成功后删除验证码
	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": verificationCode.ID})
	if err != nil {
		return false, err
	}

	return true, nil
}

// CreateIndexes 创建索引
func (s *VerificationCodeService) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.M{"name": 1, "channel": 1},
		},
		{
			Keys:    bson.M{"expires_at": 1},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
