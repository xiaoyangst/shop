package handler

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"shop_srvs/user_srv/global"
	model "shop_srvs/user_srv/model/gen"
	proto "shop_srvs/user_srv/proto"
	"time"
)

type UserServer struct {
	proto.UnimplementedUserServiceServer
}

var options = &password.Options{
	SaltLen:      16,         // 盐的长度
	Iterations:   100,        // 迭代次数
	KeyLen:       32,         // 生成的密钥长度
	HashFunction: sha512.New, // 使用 SHA-512 哈希函数
}

func ModelToResponse(User model.ListUsersRow) *proto.UserInfoResponse {
	return &proto.UserInfoResponse{
		Id:       User.ID,
		Mobile:   User.Mobile,
		Password: User.Password,
		Nickname: User.Nikename,
		Gender:   string(User.Gender.UsersGender),
		Role:     User.Role.String,
		Birthday: uint64(User.Birthday.Time.Unix()),
	}
}

func (s *UserServer) GetUserList(ctx context.Context, in *proto.PageInfo) (*proto.UserListResponse, error) {
	queries := model.New(global.DbConn)

	limit := int32(in.PageSize)
	offset := int32((in.PageIndex - 1) * in.PageSize)

	userInfos, err := queries.ListUsers(ctx, model.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, status.Error(codes.Internal, "查询用户列表失败")
	}

	total, err := queries.CountUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "查询用户总数失败")
	}

	var rsp proto.UserListResponse
	rsp.Total = int32(total)

	for _, userInfo := range userInfos {
		re := ModelToResponse(userInfo)
		rsp.Data = append(rsp.Data, re)
	}

	return &rsp, nil
}

func (s *UserServer) GetUserByMobile(ctx context.Context, in *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	queries := model.New(global.DbConn)

	userInfo, err := queries.GetUserByMobile(ctx, in.Mobile)
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	return &proto.UserInfoResponse{
		Id:       userInfo.ID,
		Mobile:   userInfo.Mobile,
		Password: userInfo.Password,
		Nickname: userInfo.Nikename,
		Gender:   string(userInfo.Gender.UsersGender),
		Role:     userInfo.Role.String,
		Birthday: uint64(userInfo.Birthday.Time.Unix()),
	}, nil
}

func (s *UserServer) GetUserById(ctx context.Context, in *proto.IdRequest) (*proto.UserInfoResponse, error) {
	queries := model.New(global.DbConn)

	userInfo, err := queries.GetUserByID(ctx, int64(in.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	return &proto.UserInfoResponse{
		Id:       userInfo.ID,
		Mobile:   userInfo.Mobile,
		Password: userInfo.Password,
		Nickname: userInfo.Nikename,
		Gender:   string(userInfo.Gender.UsersGender),
		Role:     userInfo.Role.String,
		Birthday: uint64(userInfo.Birthday.Time.Unix()),
	}, nil
}

func (s *UserServer) CreateUser(ctx context.Context, in *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	queries := model.New(global.DbConn)

	// 先确保用户不存在，避免重复创建
	_, err := queries.GetUserByMobile(ctx, in.Mobile)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "用户已存在")
	}

	// 加密密码
	pwd := global.GenPwd(in.Password)

	// 创建用户
	userInfo, err := queries.CreateUser(ctx, model.CreateUserParams{
		Mobile:   in.Mobile,
		Password: pwd,
		NikeName: in.Nickname,
		Birthday: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Gender: model.NullUsersGender{
			UsersGender: model.UsersGenderOther,
			Valid:       true,
		},
		Role: sql.NullString{
			String: "user",
			Valid:  true,
		},
	})

	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "创建用户失败")
	}

	// 得到新用户的ID
	userId, _ := userInfo.LastInsertId()

	return &proto.UserInfoResponse{
		Id:       userId,
		Mobile:   in.Mobile,
		Password: in.Password,
		Nickname: in.Nickname,
		Gender:   string(model.UsersGenderOther),
		Role:     "user",
		Birthday: uint64(time.Now().Unix()),
	}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, in *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	// 确保用户存在
	queries := model.New(global.DbConn)
	userInfo, err := queries.GetUserByMobile(ctx, in.Mobile)
	if err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	// 更新用户信息
	err = queries.UpdateUser(ctx, model.UpdateUserParams{
		ID:       userInfo.ID,
		Mobile:   in.Mobile,
		Password: in.Password,
		Nikename: in.Nickname,
		Birthday: sql.NullTime{
			Time:  time.Unix(int64(in.Birthday), 0),
			Valid: true,
		},
		Gender: model.NullUsersGender{
			UsersGender: model.UsersGender(in.Gender),
			Valid:       true,
		},
		Role: sql.NullString{
			String: in.Role,
			Valid:  true,
		},
	})

	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "更新用户失败")
	}
	return &emptypb.Empty{}, nil
}

func (s *UserServer) CheckPassword(ctx context.Context, in *proto.CheckPasswordInfo) (*proto.CheckPasswordResponse, error) {
	re := global.VerifyPwd(in.Password, in.EncryptedPassword)
	return &proto.CheckPasswordResponse{
		Success: re,
	}, nil
}
