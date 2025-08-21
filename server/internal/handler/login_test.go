package handler

import (
    "context"
    "database/sql"
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

func setupLoginTestDB(t *testing.T) *sql.DB {
	cfg := config.Config{}

	err := config.Read(&cfg)
	require.NoError(t, err, "Failed to read config")

	db, err := sql.Open("postgres", cfg.DatabaseURI)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec(storage.UsersQuery)
	require.NoError(t, err)

	return db
}

func TestGofermart_Login(t *testing.T) {
    gofakeit.Seed(0)

    db := setupLoginTestDB(t)
    defer db.Close()

    defer func() {
        _, err := db.Exec(`DROP TABLE IF EXISTS users`)
        require.NoError(t, err)
    }()

    storage := &storage.Storage{DB: db}
    service := &service.Gofermart{
        Storage: storage,
        Log:     slog.Default(),
    }

    testUser := models.User{
        Login:    gofakeit.Username(),
        Password: gofakeit.Password(true, true, true, true, false, 10),
    }

    _, err := db.Exec(`INSERT INTO users (login, password) VALUES ($1, $2)`,
        testUser.Login, testUser.Password)
    require.NoError(t, err)

    tests := []struct {
        name     string
        user     models.User
        wantErr  bool
        errCheck func(error) bool
    }{
        {
            name: "successful login",
            user: models.User{
                Login:    testUser.Login,
                Password: testUser.Password,
            },
            wantErr: false,
        },
        {
            name: "wrong password",
            user: models.User{
                Login:    testUser.Login,
                Password: gofakeit.Password(true, true, true, true, false, 12),
            },
            wantErr: true,
        },
        {
            name: "non-existent user",
            user: models.User{
                Login:    gofakeit.Username(),
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: true,
        },
        {
            name: "empty login",
            user: models.User{
                Login:    "",
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: true,
        },
        {
            name: "empty password",
            user: models.User{
                Login:    gofakeit.Username(),
                Password: "",
            },
            wantErr: true,
        },
        {
            name: "special characters in login",
            user: models.User{
                Login:    gofakeit.Username() + "!@#$%",
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: false,
        },
        {
            name: "very long login",
            user: models.User{
                Login:    gofakeit.LetterN(100),
                Password: gofakeit.Password(true, true, true, true, false, 10),
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.Login(context.Background(), tt.user)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errCheck != nil {
                    assert.True(t, tt.errCheck(err))
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}