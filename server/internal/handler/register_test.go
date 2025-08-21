package handler

import (
    "context"
    "database/sql"
    "strings"
    "testing"

    "github.com/brianvoe/gofakeit/v7"
    _ "github.com/lib/pq"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gofermart/internal/config"
    "gofermart/internal/models"
    "gofermart/internal/service"
    "gofermart/internal/storage"
    "log/slog"
)

func setupRegisterTestDB(t *testing.T) *sql.DB {
	cfg := config.Config{}

	err := config.Read(&cfg)
	require.NoError(t, err, "Failed to read config")

	db, err := sql.Open("postgres", cfg.DatabaseURI)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec(storage.UsersQuery)
	require.NoError(t, err)

	_, err = db.Exec(storage.BalancesQuery)
	require.NoError(t, err)

	return db
}

func TestGofermart_Register(t *testing.T) {
    db := setupRegisterTestDB(t)
    defer db.Close()

    defer func() {
        _, err := db.Exec(`DROP TABLE IF EXISTS users`)
        require.NoError(t, err)
        _, err = db.Exec(`DROP TABLE IF EXISTS balances`)
        require.NoError(t, err)
    }()

    storage := &storage.Storage{DB: db}
    service := &service.Gofermart{
        Storage: storage,
        Log:     slog.Default(),
    }

    baseUser := models.User{
        Login:    gofakeit.Username(),
        Password: gofakeit.Password(true, true, true, true, false, 10),
    }

    tests := []struct {
        name     string
        user     models.User
        wantErr  bool
        errCheck func(error) bool
    }{
        {
            name: "successful registration with random data",
            user: models.User{
                Login:    gofakeit.Username(),
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: false,
        },
        {
            name:    "duplicate user",
            user:    baseUser,
            wantErr: true,
            errCheck: func(err error) bool {
                return strings.Contains(err.Error(), "duplicate key value")
            },
        },
        {
            name: "random email as login",
            user: models.User{
                Login:    gofakeit.Email(),
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: false,
        },
        {
            name: "very long login",
            user: models.User{
                Login:    gofakeit.LetterN(50),
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: false,
        },
        {
            name: "special characters in login",
            user: models.User{
                Login:    gofakeit.Username() + "@#$%",
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: false,
        },
        {
            name: "very long password",
            user: models.User{
                Login:    gofakeit.Username(),
                Password: gofakeit.Password(true, true, true, true, false, 100),
            },
            wantErr: false,
        },
    }

    err := service.Register(context.Background(), baseUser)
    require.NoError(t, err)

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.Register(context.Background(), tt.user)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errCheck != nil {
                    assert.True(t, tt.errCheck(err))
                }
            } else {
                assert.NoError(t, err)

                var storedUser models.User
                err = db.QueryRow("SELECT login, password FROM users WHERE login = $1",
                    tt.user.Login).Scan(&storedUser.Login, &storedUser.Password)
                assert.NoError(t, err)
                assert.Equal(t, tt.user.Login, storedUser.Login)
                assert.Equal(t, tt.user.Password, storedUser.Password)

                var balance float64
                var withdrawn float64
                err = db.QueryRow("SELECT current, withdrawn FROM balances WHERE login = $1",
                    tt.user.Login).Scan(&balance, &withdrawn)
                assert.NoError(t, err)
                assert.Equal(t, float64(0), balance)
                assert.Equal(t, float64(0), withdrawn)
            }
        })
    }
}
