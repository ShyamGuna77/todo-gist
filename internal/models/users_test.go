package models

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHashArg struct {
	password string
}

func (a bcryptHashArg) Match(v driver.Value) bool {
	switch x := v.(type) {
	case string:
		return bcrypt.CompareHashAndPassword([]byte(x), []byte(a.password)) == nil
	case []byte:
		return bcrypt.CompareHashAndPassword(x, []byte(a.password)) == nil
	default:
		return false
	}
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestUserModelInsert_HashesPassword(t *testing.T) {
	db, mock := newMockDB(t)
	m := &UserModel{DB: db}

	const (
		name     = "Alice"
		email    = "alice@example.com"
		password = "correct horse battery staple"
	)

	stmt := "INSERT INTO users (name, email, hashed_password, created)\n    VALUES(?, ?, ?, UTC_TIMESTAMP())"

	mock.ExpectExec(regexp.QuoteMeta(stmt)).
		WithArgs(name, email, bcryptHashArg{password: password}).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := m.Insert(name, email, password); err != nil {
		t.Fatalf("Insert() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestUserModelInsert_DuplicateEmail(t *testing.T) {
	db, mock := newMockDB(t)
	m := &UserModel{DB: db}

	const (
		name     = "Bob"
		email    = "bob@example.com"
		password = "supersecretpassword"
	)

	stmt := "INSERT INTO users (name, email, hashed_password, created)\n    VALUES(?, ?, ?, UTC_TIMESTAMP())"

	mock.ExpectExec(regexp.QuoteMeta(stmt)).
		WithArgs(name, email, sqlmock.AnyArg()).
		WillReturnError(&mysql.MySQLError{
			Number:  1062,
			Message: "Duplicate entry 'bob@example.com' for key 'users_uc_email'",
		})

	err := m.Insert(name, email, password)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestUserModelAuthenticate_Success(t *testing.T) {
	db, mock := newMockDB(t)
	m := &UserModel{DB: db}

	const (
		email    = "carol@example.com"
		password = "verylongpassword"
	)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword(): %v", err)
	}

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"
	rows := sqlmock.NewRows([]string{"id", "hashed_password"}).AddRow(42, []byte(hash))

	mock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(email).WillReturnRows(rows)

	id, err := m.Authenticate(email, password)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if id != 42 {
		t.Fatalf("Authenticate() id = %d, want %d", id, 42)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestUserModelAuthenticate_InvalidCredentials_WrongPassword(t *testing.T) {
	db, mock := newMockDB(t)
	m := &UserModel{DB: db}

	const (
		email        = "dave@example.com"
		rightPass    = "rightpassword"
		wrongPass    = "wrongpassword"
		expectedUser = 7
	)

	hash, err := bcrypt.GenerateFromPassword([]byte(rightPass), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword(): %v", err)
	}

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"
	rows := sqlmock.NewRows([]string{"id", "hashed_password"}).AddRow(expectedUser, []byte(hash))
	mock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(email).WillReturnRows(rows)

	id, err := m.Authenticate(email, wrongPass)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if id != 0 {
		t.Fatalf("expected id 0 on invalid creds, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestUserModelAuthenticate_InvalidCredentials_NoRows(t *testing.T) {
	db, mock := newMockDB(t)
	m := &UserModel{DB: db}

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"
	mock.ExpectQuery(regexp.QuoteMeta(stmt)).
		WithArgs("nobody@example.com").
		WillReturnError(sql.ErrNoRows)

	id, err := m.Authenticate("nobody@example.com", "whatever")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if id != 0 {
		t.Fatalf("expected id 0 on invalid creds, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
