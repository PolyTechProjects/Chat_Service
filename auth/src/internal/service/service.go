package service

import (
	"fmt"
	"time"

	"example.com/main/src/internal/client"
	"example.com/main/src/internal/dto"
	"example.com/main/src/internal/mail"
	"example.com/main/src/internal/repository"
	"example.com/main/src/models"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

var mailsXVerificationType = map[string]mail.Mail{
	"email":       mail.NewEmailVerificationMail(),
	"phone":       mail.NewPasswordRestoreMail(),
	"2fa_connect": mail.NewTwoFactorAuthConnectMail(),
	"2fa_code":    mail.NewTwoFactorAuthCodeMail(),
}

type AuthService struct {
	AuthRepository *repository.AuthRepository
	KeycloakClient *client.KeycloakClient
	RedisClient    *client.RedisClient
	totpSecret     string
}

func New(authRepository *repository.AuthRepository, keycloakClient *client.KeycloakClient, redisClient *client.RedisClient, totpSecret string) *AuthService {
	return &AuthService{
		AuthRepository: authRepository,
		KeycloakClient: keycloakClient,
		RedisClient:    redisClient,
		totpSecret:     totpSecret,
	}
}

func (s *AuthService) Register(login string, username string, password string, firstname string, lastname string) (*models.User, error) {
	user, err := models.NewAccount(login, username, password, firstname, lastname)
	if err != nil {
		return nil, err
	}
	err = s.KeycloakClient.RegisterUser(user)
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Pass = string(hash)
	err = s.AuthRepository.Save(user)
	if err != nil {
		return nil, err
	}
	accountCreatedEvent := &dto.AccountCreatedEvent{
		UserId:    user.Id.String(),
		Login:     user.Login,
		Username:  user.Name,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
	}
	s.RedisClient.SendToCreateAccountChannel(accountCreatedEvent)
	s.SendVerificationMessage("email", user.Login)

	return user, nil
}

func (s *AuthService) Login(login string, password string) (string, string, error) {
	user, err := s.AuthRepository.FindByLogin(login)
	if err != nil {
		return "", "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(password))
	if err != nil {
		return "", "", err
	}

	return s.KeycloakClient.LoginUser(login, password)
}

func (s *AuthService) Authorize(accessToken string) error {
	return s.KeycloakClient.AuthroizeUser(accessToken)
}

func (s *AuthService) RefreshTokens(refreshToken string) (string, string, error) {
	return s.KeycloakClient.RefreshTokens(refreshToken)
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.KeycloakClient.RevokeTokens(refreshToken)
}

func (s *AuthService) DeleteAccount(userId uuid.UUID, accessToken string, refreshToken string) error {
	user, err := s.AuthRepository.FindById(userId)
	if err != nil {
		return err
	}
	err = s.KeycloakClient.DeleteAccount(user.KeycloakId, accessToken, refreshToken)
	if err != nil {
		return err
	}
	err = s.AuthRepository.DeleteById(userId)
	if err != nil {
		return err
	}
	accountDeletedEvent := &dto.AccountDeletedEvent{
		UserId: user.Id.String(),
	}
	s.RedisClient.SendToDeleteAccountChannel(accountDeletedEvent)

	return nil
}

func (s *AuthService) ExtractUserId(accessToken string) (string, error) {
	keycloak, err := s.KeycloakClient.ExtractUserId(accessToken)
	if err != nil {
		return "", err
	}
	keycloakId, err := uuid.Parse(keycloak)
	if err != nil {
		return "", err
	}
	user, err := s.AuthRepository.FindByKeycloakId(keycloakId)
	if err != nil {
		return "", err
	}
	return user.Id.String(), nil
}

func (s *AuthService) GetLogin(userId string) (string, error) {
	id, err := uuid.Parse(userId)
	if err != nil {
		return "", err
	}
	user, err := s.AuthRepository.FindById(id)
	if err != nil {
		return "", err
	}
	return user.Login, nil
}

func (s *AuthService) SendVerificationMessage(verificationType string, email string) error {
	verificationToken, err := s.startUserVerification(email)
	if err != nil {
		return err
	}

	mailBody := mailsXVerificationType[verificationType].BuildBody(mail.MailBodyOpts{
		VerificationToken: verificationToken,
	})
	sendMailEvent := mailsXVerificationType[verificationType].Build(email, mailBody)
	s.RedisClient.SendToSendMailChannel(sendMailEvent)

	return nil
}

func (s *AuthService) VerifyEmail(req *dto.VerifyEmailRequest) error {
	userVerification, err := s.verifyUser(req.VerificationToken)
	if err != nil {
		return err
	}
	user, err := s.AuthRepository.FindById(userVerification.UserId)
	if err != nil {
		return err
	}
	user.IsVerified = true
	err = s.AuthRepository.Save(user)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ResetPassword(req *dto.ResetPasswordRequest) error {
	userVerification, err := s.verifyUser(req.VerificationToken)
	if err != nil {
		return err
	}
	user, err := s.AuthRepository.FindById(userVerification.UserId)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Pass = string(hash)
	err = s.AuthRepository.Save(user)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ChangePassword(req *dto.ChangePasswordRequest, accessToken string) error {
	userID, err := s.KeycloakClient.ExtractUserId(accessToken)
	if err != nil {
		return err
	}
	userId, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	user, err := s.AuthRepository.FindById(userId)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(req.OldPassword))
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Pass = string(hash)
	err = s.AuthRepository.Save(user)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) Connect2FA(accessToken string) error {
	userID, err := s.KeycloakClient.ExtractUserId(accessToken)
	if err != nil {
		return err
	}
	userId, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	user, err := s.AuthRepository.FindById(userId)
	if err != nil {
		return err
	}
	user.Is2FAEnabled = true
	err = s.AuthRepository.Save(user)
	if err != nil {
		return err
	}

	twoFactorAuthCode, err := s.startUserVerification(user.Login)
	if err != nil {
		return err
	}
	mailBody := mailsXVerificationType["2fa_connect"].BuildBody(mail.MailBodyOpts{
		TwoFactorAuthCode: twoFactorAuthCode,
	})
	sendMailEvent := mailsXVerificationType["2fa_connect"].Build(user.Login, mailBody)
	s.RedisClient.SendToSendMailChannel(sendMailEvent)
	return nil
}

func (s *AuthService) verifyUser(verificationToken string) (*models.UserVerification, error) {
	userVerification, err := s.AuthRepository.FindUserVerificationByToken(verificationToken)
	if err != nil {
		return nil, err
	}
	if userVerification.IsExpired {
		return nil, fmt.Errorf("verification token is expired")
	}
	if ok := totp.Validate(userVerification.VerificationToken, s.totpSecret); !ok {
		return nil, fmt.Errorf("verification token is invalid")
	}
	userVerification.IsExpired = true
	_, err = s.AuthRepository.StoreUserVerification(userVerification)
	if err != nil {
		return nil, err
	}
	return userVerification, nil
}

func (s *AuthService) startUserVerification(email string) (string, error) {
	code, err := totp.GenerateCode(s.totpSecret, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to generate totp. error=%w", err)
	}

	user, err := s.AuthRepository.FindByLogin(email)
	if err != nil {
		return "", err
	}
	userVerification := models.NewUserVerification(user.Id, code)
	userVerificationOld, err := s.AuthRepository.FindUserVerificationById(user.Id)
	if err != nil {
		return "", err
	}
	if userVerificationOld != nil {
		userVerificationOld.IsExpired = true
		_, err = s.AuthRepository.StoreUserVerification(userVerificationOld)
		if err != nil {
			return "", err
		}
	}
	_, err = s.AuthRepository.StoreUserVerification(userVerification)
	if err != nil {
		return "", err
	}
	return code, err
}
