package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"godas/model/domain"
	"godas/model/web"
	"godas/repository"
	"math/rand"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/go-playground/validator/v10"
)

type EmailVerificationService interface {
	Create(domain.EmailVerificationSend) (domain.EmailVerification, error)
	Recreate(domain.EmailVerificationSend) (domain.EmailVerification, error)
	Verification(web.EmailVerificationCreateRequest) error
}

type EmailVerificationServiceImpl struct {
	emailVerificationRepository repository.EmailVerificationRepository
	validate                    *validator.Validate
}

func NewEmailVerificationService(emailVerificationRepository repository.EmailVerificationRepository, validate *validator.Validate) EmailVerificationService {
	return &EmailVerificationServiceImpl{
		emailVerificationRepository: emailVerificationRepository,
		validate:                    validate,
	}
}

var (
	codeSelection       = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	codeSelectionLength = len(codeSelection)
)

func (service *EmailVerificationServiceImpl) GenerateCode() []byte {
	code := make([]byte, 6)
	codeLength := len(code)
	for i := 0; i < codeLength; i++ {
		code[i] = codeSelection[rand.Intn(codeSelectionLength)]
	}
	return code
}

func (service *EmailVerificationServiceImpl) Create(email domain.EmailVerificationSend) (domain.EmailVerification, error) {
	emailVerification := domain.EmailVerification{}

	from := mail.Address{
		Name:    email.FromName,
		Address: email.FromEmail,
	}

	to := mail.Address{
		Address: email.ToEmail,
	}

	header := map[string]string{
		"From":                      from.String(),
		"To":                        to.String(),
		"Subject":                   email.Title,
		"MIME-Version":              "1.0",
		"Content-Type":              "text/plain; charset=\"utf-8\"",
		"Content-Transfer-Encoding": "base64",
	}

	message := ""
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	code := service.GenerateCode()
	message += "\r\n" + base64.StdEncoding.EncodeToString(code)

	auth := smtp.PlainAuth("", email.FromEmail, email.FromPassword, email.Host)

	if err := smtp.SendMail(
		fmt.Sprintf("%v:%v", email.Host, email.Port),
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	); err != nil {
		return emailVerification, err
	}

	emailVerification = domain.EmailVerification{
		Email:      email.ToEmail,
		Code:       string(code),
		Expiration: time.Now().Add(time.Minute * 10).Unix(),
		Cooldown:   time.Now().Add(time.Minute).Unix(),
	}

	if err := service.emailVerificationRepository.Insert(context.Background(), emailVerification); err != nil {
		if errors.Is(err, repository.ErrDuplicateData) {
			return emailVerification, ErrDuplicate
		}
		return emailVerification, err
	}

	return emailVerification, nil
}

func (service *EmailVerificationServiceImpl) Recreate(email domain.EmailVerificationSend) (domain.EmailVerification, error) {
	emailVerification, err := service.emailVerificationRepository.FindByEmail(context.Background(), email.ToEmail)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return emailVerification, ErrNotFound
		}
		return emailVerification, err
	}

	if time.Now().Unix() < emailVerification.Cooldown {
		return emailVerification, ErrUnauthorized
	}

	emailVerification.Expiration = time.Now().Add(time.Minute * 10).Unix()
	emailVerification.Cooldown = time.Now().Add(time.Minute * 1).Unix()

	from := mail.Address{
		Name:    email.FromName,
		Address: email.FromEmail,
	}

	to := mail.Address{
		Address: email.ToEmail,
	}

	header := map[string]string{
		"From":                      from.String(),
		"To":                        to.String(),
		"Subject":                   email.Title,
		"MIME-Version":              "1.0",
		"Content-Type":              "text/plain; charset=\"utf-8\"",
		"Content-Transfer-Encoding": "base64",
	}

	message := ""
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	code := service.GenerateCode()
	message += "\r\n" + base64.StdEncoding.EncodeToString(code)
	emailVerification.Code = string(code)

	auth := smtp.PlainAuth("", email.FromEmail, email.FromPassword, email.Host)

	if err := smtp.SendMail(
		fmt.Sprintf("%v:%v", email.Host, email.Port),
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	); err != nil {
		return emailVerification, err
	}

	res, err := service.emailVerificationRepository.Update(context.Background(), emailVerification)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateData) {
			return emailVerification, ErrDuplicate
		} else if errors.Is(err, repository.ErrNoData) {
			return emailVerification, ErrNotFound
		}
		return emailVerification, err
	}

	return res, nil
}

func (service *EmailVerificationServiceImpl) Verification(request web.EmailVerificationCreateRequest) error {
	if err := service.validate.Struct(request); err != nil {
		return ErrBadRequest
	}

	emailVerification, err := service.emailVerificationRepository.FindByEmail(context.Background(), request.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return ErrNotFound
		}
		return err
	}

	if time.Now().Unix() >= emailVerification.Expiration {
		return ErrUnauthorized
	}

	if request.Code != emailVerification.Code {
		return ErrUnauthorized
	}

	if err := service.emailVerificationRepository.Delete(context.Background(), request.Email); err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return ErrNotFound
		}
		return err
	}

	return nil
}
