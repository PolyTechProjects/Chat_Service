package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"example.com/main/src/config"
	"example.com/main/src/internal/dto"
	"example.com/main/src/models"
	"github.com/Nerzal/gocloak/v13"
	"github.com/go-redis/redis"
)

type KeycloakClient struct {
	client       *gocloak.GoCloak
	realm        string
	clientId     string
	clientSecret string
}

func NewKeycloakClient(cfg *config.Config) *KeycloakClient {
	keycloakUrl := fmt.Sprintf("http://%s:%d", cfg.Keycloak.Host, cfg.Keycloak.InnerPort)
	slog.Info("Keycloak url: " + keycloakUrl)
	slog.Info("Keycloak realm: " + cfg.Keycloak.Realm)
	client := gocloak.NewClient(keycloakUrl)
	return &KeycloakClient{client: client, realm: cfg.Keycloak.Realm, clientId: cfg.Keycloak.ClientId, clientSecret: cfg.Keycloak.ClientSecret}
}

func (k *KeycloakClient) adminAuth() *gocloak.JWT {
	token, err := k.client.LoginClient(context.Background(), k.clientId, k.clientSecret, k.realm)
	if err != nil {
		panic("Login failed: " + err.Error())
	}
	return token
}

func (k *KeycloakClient) RegisterUser(user *models.User) error {
	token := k.adminAuth()
	keycloakUser := gocloak.User{
		ID:        gocloak.StringP(user.Id.String()),
		Username:  gocloak.StringP(user.Login),
		Email:     gocloak.StringP(user.Login),
		Enabled:   gocloak.BoolP(true),
		FirstName: gocloak.StringP(user.Firstname),
		LastName:  gocloak.StringP(user.Lastname),
	}
	keycloakUserId, err := k.client.CreateUser(context.Background(), token.AccessToken, k.realm, keycloakUser)
	if err != nil {
		slog.Error("KeycloakCreateUser failed: " + err.Error())
		return err
	}
	user.KeycloakId = keycloakUserId
	err = k.client.SetPassword(context.Background(), token.AccessToken, user.KeycloakId, k.realm, user.Pass, false)
	if err != nil {
		slog.Error("KeycloakSetPassword failed: " + err.Error())
		return err
	}
	return nil
}

func (k *KeycloakClient) LoginUser(login string, password string) (string, string, error) {
	token, err := k.client.Login(context.Background(), k.clientId, k.clientSecret, k.realm, login, password)
	if err != nil {
		slog.Error("KeycloakLogin failed: " + err.Error())
		return "", "", err
	}
	return token.AccessToken, token.RefreshToken, nil
}

func (k *KeycloakClient) AuthroizeUser(accessToken string) error {
	res, err := k.client.RetrospectToken(context.Background(), accessToken, k.clientId, k.clientSecret, k.realm)
	if err != nil {
		slog.Error("KeycloakRetrospectToken failed: " + err.Error())
		return err
	}
	if !*res.Active {
		slog.Error("Token is not active")
		return fmt.Errorf("token is not active")
	}
	return nil
}

func (k *KeycloakClient) RefreshTokens(refreshToken string) (string, string, error) {
	token, err := k.client.RefreshToken(context.Background(), refreshToken, k.clientId, k.clientSecret, k.realm)
	if err != nil {
		slog.Error("KeycloakRefreshToken failed: " + err.Error())
		return "", "", err
	}
	return token.AccessToken, token.RefreshToken, nil
}

func (k *KeycloakClient) RevokeTokens(refreshToken string) error {
	err := k.client.RevokeToken(context.Background(), k.realm, k.clientId, k.clientSecret, refreshToken)
	if err != nil {
		slog.Error("KeycloakRevokeToken failed: " + err.Error())
		return err
	}
	return nil
}

func (k *KeycloakClient) DeleteAccount(userId string, accessToken string, refreshToken string) error {
	token := k.adminAuth()
	k.client.GetUserByID(context.Background(), token.AccessToken, k.realm, userId)
	res, err := k.client.RetrospectToken(context.Background(), accessToken, k.clientId, k.clientSecret, k.realm)
	if err != nil {
		slog.Error("KeycloakRetrospectToken failed: " + err.Error())
		return err
	}
	if !*res.Active {
		slog.Error("Token is not active")
		return fmt.Errorf("token is not active")
	}
	_, claims, err := k.client.DecodeAccessToken(context.Background(), accessToken, k.realm)
	if err != nil {
		slog.Error("KeycloakDecodeAccessToken failed: " + err.Error())
		return err
	}
	subject, err := claims.GetSubject()
	if err != nil {
		slog.Error("KeycloakGetSubject failed: " + err.Error())
		return err
	}
	if subject != userId {
		slog.Error("Invalid user id")
		return fmt.Errorf("invalid user id")
	}
	err = k.RevokeTokens(refreshToken)
	if err != nil {
		slog.Error("KeycloakRevokeToken failed: " + err.Error())
		return err
	}
	return k.client.DeleteUser(context.Background(), token.AccessToken, k.realm, userId)
}

type RedisClient struct {
	Client                   *redis.Client
	createAccountChannelName string
	deleteAccountChannelName string
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.InnerPort),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})
	return &RedisClient{
		Client:                   client,
		createAccountChannelName: cfg.Redis.CreateAccountChannelName,
		deleteAccountChannelName: cfg.Redis.DeleteAccountChannelName,
	}
}

func (c *RedisClient) SendToCreateAccountChannel(accountCreatedEvent *dto.AccountCreatedEvent) error {
	event, err := json.Marshal(accountCreatedEvent)
	if err != nil {
		slog.Error("Failed to marshal event: " + err.Error())
		return err
	}
	_, err = c.Client.Publish(c.createAccountChannelName, event).Result()
	if err != nil {
		slog.Error("Failed to publish message: " + err.Error())
		return err
	}
	return nil
}

func (c *RedisClient) SendToDeleteAccountChannel(accountDeletedEvent *dto.AccountDeletedEvent) error {
	event, err := json.Marshal(accountDeletedEvent)
	if err != nil {
		slog.Error("Failed to marshal event: " + err.Error())
		return err
	}
	_, err = c.Client.Publish(c.deleteAccountChannelName, event).Result()
	if err != nil {
		slog.Error("Failed to publish message: " + err.Error())
		return err
	}
	return nil
}
