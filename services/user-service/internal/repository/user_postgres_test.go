package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) (*UserRepository, pgxmock.PgxPoolIface) {
	t.Helper()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	t.Cleanup(func() { mockPool.Close() })
	return NewUserRepository(mockPool), mockPool
}

func TestUserRepository_Create(t *testing.T) {
	repo, mock := newTestRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (id, email, password)")).
		WithArgs(pgxmock.AnyArg(), "alice@example.com", "hash").
		WillReturnRows(pgxmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

	user, err := repo.Create("alice@example.com", "hash")

	require.NoError(t, err)
	require.Equal(t, "alice@example.com", user.Email)
	require.NotEmpty(t, user.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CreateDuplicate(t *testing.T) {
	repo, mock := newTestRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (id, email, password)")).
		WithArgs(pgxmock.AnyArg(), "dup@example.com", "hash").
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err := repo.Create("dup@example.com", "hash")

	require.ErrorIs(t, err, ErrUserExists)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo, mock := newTestRepo(t)
	created := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, created_at FROM users")).
		WithArgs("alice@example.com").
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password", "created_at"}).AddRow("id-1", "alice@example.com", "hash", created))

	user, err := repo.GetByEmail("alice@example.com")
	require.NoError(t, err)
	require.Equal(t, "id-1", user.ID)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, created_at FROM users")).
		WithArgs("missing@example.com").
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByEmail("missing@example.com")
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserRepository_UpdateAndDelete(t *testing.T) {
	repo, mock := newTestRepo(t)

	newEmail := "new@example.com"
	newPass := "newpass"
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WithArgs(&newEmail, &newPass, "user-1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	require.NoError(t, repo.UpdateFields("user-1", &newEmail, &newPass))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WithArgs((*string)(nil), (*string)(nil), "user-2").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	require.ErrorIs(t, repo.UpdateFields("user-2", nil, nil), ErrUserNotFound)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users")).
		WithArgs("user-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	require.NoError(t, repo.Delete("user-1"))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users")).
		WithArgs("user-3").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	require.ErrorIs(t, repo.Delete("user-3"), ErrUserNotFound)
}
