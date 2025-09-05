package db

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User 用户结构体
// @Description 用户信息结构体
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // 用户ID
	Name      string             `bson:"name" json:"name"`                  // 用户名，唯一键
	QQ        string             `bson:"qq" json:"qq"`                      // QQ号
	Phone     string             `bson:"phone" json:"phone"`                // 手机号
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`      // 创建时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`      // 更新时间
}

// UserService 用户服务
type UserService struct {
	collection *mongo.Collection
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		collection: Collection("message_db", "users"),
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
	// 检查用户名是否已存在
	filter := bson.M{"name": user.Name}
	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户名已存在")
	}

	// 设置时间戳
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 插入用户
	result, err := s.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	// 获取插入的ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}

	return nil
}

// GetUserByName 根据用户名获取用户
func (s *UserService) GetUserByName(ctx context.Context, name string) (*User, error) {
	var user User
	filter := bson.M{"name": name}
	err := s.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id string) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("无效的ID格式")
	}

	var user User
	filter := bson.M{"_id": objectID}
	err = s.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, id string, updateData map[string]interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("无效的ID格式")
	}

	// 检查用户名是否被其他用户使用
	if name, ok := updateData["name"].(string); ok {
		filter := bson.M{"name": name, "_id": bson.M{"$ne": objectID}}
		count, err := s.collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("用户名已被其他用户使用")
		}
	}

	// 设置更新时间
	updateData["updated_at"] = time.Now()

	update := bson.M{"$set": updateData}
	filter := bson.M{"_id": objectID}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("用户不存在")
	}

	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("无效的ID格式")
	}

	filter := bson.M{"_id": objectID}
	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("用户不存在")
	}

	return nil
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUsersByPage 分页获取用户
func (s *UserService) GetUsersByPage(ctx context.Context, page, pageSize int64) ([]User, int64, error) {
	var users []User
	skip := (page - 1) * pageSize

	// 获取总数
	total, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(pageSize).
		SetSort(bson.M{"created_at": -1})

	cursor, err := s.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// CreateIndexes 创建索引
func (s *UserService) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.M{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"qq": 1},
		},
		{
			Keys: bson.M{"phone": 1},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
