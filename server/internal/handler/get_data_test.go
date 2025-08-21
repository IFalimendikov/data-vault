package handler

import (
    "context"
    "database/sql"
    "testing"
    "sync"
    
    "gofermart/internal/config"
    "gofermart/internal/models"
    "gofermart/internal/service"
    "gofermart/internal/storage"
    
    "github.com/brianvoe/gofakeit/v7"
    "github.com/ShiraazMoollatjie/goluhn"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupGetOrdersTestDB(t *testing.T) *sql.DB {
	cfg := config.Config{}
	
	err := config.Read(&cfg)
	require.NoError(t, err, "Failed to read config")
	db, err := sql.Open("postgres", cfg.DatabaseURI)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec(storage.OrdersQuery)
	require.NoError(t, err)
	_, err = db.Exec(storage.UsersQuery)
	require.NoError(t, err)
	_, err = db.Exec(storage.BalancesQuery)
	require.NoError(t, err)
	return db
}

func TestGofermart_ConcurrentBalanceOperations(t *testing.T) {
    gofakeit.Seed(0)

    db := setupGetOrdersTestDB(t)
    defer db.Close()
    defer func() {
        _, err := db.Exec(`DROP TABLE IF EXISTS orders`)
        require.NoError(t, err)
        _, err = db.Exec(`DROP TABLE IF EXISTS users`)
        require.NoError(t, err)
        _, err = db.Exec(`DROP TABLE IF EXISTS balances`)
        require.NoError(t, err)
    }()

    storage := &storage.Storage{DB: db}
    service := &service.Gofermart{
        Storage: storage,
    }

    testUser := models.User{
        Login:    gofakeit.Username(),
        Password: gofakeit.Password(true, true, true, true, false, 10),
    }

    err := service.Register(context.Background(), testUser)
    require.NoError(t, err, "Failed to register test user")

    initialBalance := gofakeit.Float64Range(1000, 5000)

    _, err = db.Exec(`
        INSERT INTO balances (login, current, withdrawn)
        VALUES ($1, $2, $3)
        ON CONFLICT (login) DO UPDATE 
        SET current = $2, withdrawn = $3
    `, testUser.Login, initialBalance, 0.0)
    require.NoError(t, err)

    var wg sync.WaitGroup
    concurrentOperations := gofakeit.Number(5, 20)
    wg.Add(concurrentOperations)

    accrualAmount := gofakeit.Float64Range(10, 100)
    withdrawAmount := gofakeit.Float64Range(5, 50)

    expectedAccrual := 0.0
    expectedWithdrawn := 0.0

    for i := 0; i < concurrentOperations; i++ {
        go func(i int) {
            defer wg.Done()

            tx, err := db.Begin()
            require.NoError(t, err)

            defer func() {
                if err != nil {
                    tx.Rollback()
                }
            }()

            if i%2 == 0 {
                orderNum := goluhn.Generate(8)
                _, err := tx.Exec(`
                    INSERT INTO orders (number, login, status, accrual, uploaded_at)
                    VALUES ($1, $2, $3, $4, NOW())
                    ON CONFLICT (number) DO NOTHING
                `, orderNum, testUser.Login, "PROCESSED", accrualAmount)
                require.NoError(t, err)

                expectedAccrual += accrualAmount

                _, err = tx.Exec(`
                    UPDATE balances 
                    SET current = current + $1
                    WHERE login = $2
                `, accrualAmount, testUser.Login)
                require.NoError(t, err)
            } else {
                expectedWithdrawn += withdrawAmount
                _, err := tx.Exec(`
                    UPDATE balances 
                    SET current = current - $1,
                        withdrawn = withdrawn + $1
                    WHERE login = $2
                `, withdrawAmount, testUser.Login)
                require.NoError(t, err)
            }

            err = tx.Commit()
            require.NoError(t, err)
        }(i)
    }

    wg.Wait()

    var currentBalance, withdrawnBalance float64
    err = db.QueryRow(`
        SELECT current, withdrawn FROM balances WHERE login = $1
    `, testUser.Login).Scan(&currentBalance, &withdrawnBalance)
    require.NoError(t, err)

    expectedFinalBalance := initialBalance + expectedAccrual - expectedWithdrawn
    assert.Equal(t, expectedFinalBalance, currentBalance, "Unexpected final current balance")
    assert.Equal(t, expectedWithdrawn, withdrawnBalance, "Unexpected final withdrawn balance")
}