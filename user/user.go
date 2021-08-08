// Package user contains all the functionality that relates to user management
package user

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Type string

var PasswordAuth Type = "password"
var EmailAuth Type = "passwordless-email"
var GoogleAuth Type = "google"

var ErrEmailInvalid = errors.New("Email is invalid")
var ErrPasswordShort = errors.New("Password cannot be shorter than 8 characters")
var ErrPasswordLong = errors.New("Password cannot be longer than 120 characters")
var ErrPasswordInvalid = errors.New("Password requires at least 1 capital letter, 1 small letter and a number")
var ErrSamePassword = errors.New("Current password already in use. Please pick another password")
var ErrActivationTokenInvalid = errors.New("Activation Token is invalid")
var ErrForgetPasswordTokenInvalid = errors.New("Forget Password Token is invalid")

// User represents a single user - current assumes that user would be one that a google account
type User struct {
	ID                       string    `gorm:"type:varchar(40);primary_key"`
	Email                    string    `gorm:"type:varchar(250)"`
	Password                 string    `json:"-" gorm:"type:varchar(250)"`
	ForgetPasswordToken      string    `json:"-" gorm:"type:varchar(40)"`
	ForgetPasswordExpiryDate time.Time `json:"-"`
	ActivationToken          string    `json:"-" gorm:"type:varchar(40)"`
	ActivationExpiryDate     time.Time `json:"-"`
	Activated                bool      `json:"-"`
	// Google Auth Tokens
	RefreshToken string `gorm:"type:varchar(250)"`
	AuthToken    string `gorm:"type:varchar(250)"`
	Type         string `gorm:"type:varchar(250)"`
	DateCreated  time.Time
	DateModified time.Time
}

func NewUser(firstName, lastName, email, password string) (*User, error) {
	user := User{Email: email}
	err := user.setPassword(password)
	user.ID = uuid.New().String()
	user.ActivationToken = uuid.New().String()
	user.DateCreated = time.Now()
	user.DateModified = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	user.ForgetPasswordExpiryDate = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	user.ActivationExpiryDate = time.Now().Add(1 * time.Hour)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u User) validateEmail() error {
	reEmail := regexp.MustCompile(`\w+@\w{2,3}.\w{2,3}`)
	isValid := reEmail.MatchString(u.Email)
	if isValid {
		return nil
	}
	return ErrEmailInvalid
}

func (u *User) setPassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordShort
	}
	if len(password) > 120 {
		return ErrPasswordLong
	}
	reSmallLetters := regexp.MustCompile("[a-z]")
	reCapital := regexp.MustCompile("[A-Z]")
	reNumbers := regexp.MustCompile("[0-9]")
	smallLettersFind := reSmallLetters.FindAllString(password, -1)
	capitalFind := reCapital.FindAllString(password, -1)
	numberFind := reNumbers.FindAllString(password, -1)
	if len(smallLettersFind) > 0 && len(capitalFind) > 0 && len(numberFind) > 0 {
		hashedPassword, errEncrpt := bcrypt.GenerateFromPassword([]byte(password), 10)
		if errEncrpt != nil {
			return ErrPasswordInvalid
		}
		u.Password = string(hashedPassword)
		return nil
	}
	return ErrPasswordInvalid
}

func (u *User) IsPasswordCorrect(password string) bool {
	parsedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return false
	}
	if u.Password == string(parsedPassword) {
		return true
	}
	return false
}

// ForgetPassword resets the forget password token to a random UUID as well as resets the
// forget password expiry token. The function will return the forgetPasswordToken
func (u *User) ForgetPassword() (string, error) {
	u.ForgetPasswordToken = uuid.New().String()
	u.ForgetPasswordExpiryDate = time.Now()
	return u.ForgetPasswordToken, nil
}

// ChangePasswordFromForget requires you to provide the forget password token. This function
// will then check the forgetPasswordToken if its correct and alters it accordingly
func (u *User) ChangePasswordFromForget(forgetPasswordToken, password string) error {
	if u.ForgetPasswordToken == forgetPasswordToken {
		err := u.setPassword(password)
		if err != nil {
			return err
		}
		return nil
	}
	return ErrForgetPasswordTokenInvalid
}

// ChangePassword changes the password on the user object before saving it
func (u *User) ChangePassword(password string) error {
	errCompare := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if errCompare == nil {
		return ErrSamePassword
	}
	err := u.setPassword(password)
	if err != nil {
		return err
	}
	return nil
}

// ReactivateToken resets the activation token in the case the user did not activate the account
// in time. Returns an activationToken
func (u *User) ReactivateToken() (string, error) {
	newToken := uuid.New().String()
	u.ActivationToken = newToken
	return newToken, nil
}

// Activate user. You would need to provide a activation token to check if it correct.
// If correct, it would return the status of the user which should be true or false
func (u *User) Activate(activationToken string) (bool, error) {
	if u.ActivationToken == activationToken {
		u.Activated = true
		return true, nil
	}
	return false, ErrActivationTokenInvalid
}
