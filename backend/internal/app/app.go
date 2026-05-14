package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"ai-interview/backend/internal/config"
	"ai-interview/backend/internal/handler"
	mysqlrepo "ai-interview/backend/internal/repository/mysql"
	"ai-interview/backend/internal/service"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type App struct {
	config config.Config
}

func New() *App {
	return &App{
		config: config.Load(),
	}
}

func (a *App) Run() error {
	db, err := openMySQL(a.config.MySQL)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := mysqlrepo.NewInterviewRepository(db)
	providerFactory := service.NewProviderFactory(a.config)
	svc := service.NewInterviewService(repo, providerFactory)
	apiHandler := handler.NewAPIHandler(svc)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.config.Server.Port),
		Handler:           apiHandler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("api server listening on http://localhost:%d\n", a.config.Server.Port)
	fmt.Printf("default ai provider: %s\n", providerFactory.Default().Name())
	return server.ListenAndServe()
}

func openMySQL(cfg config.MySQLConfig) (*sql.DB, error) {
	dsn := cfg.DSN
	if dsn == "" {
		driverConfig := mysqlDriver.NewConfig()
		driverConfig.Net = "tcp"
		driverConfig.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		driverConfig.User = cfg.User
		driverConfig.Passwd = cfg.Password
		driverConfig.DBName = cfg.Database
		driverConfig.ParseTime = true
		driverConfig.Params = map[string]string{
			"charset": "utf8mb4",
		}
		dsn = driverConfig.FormatDSN()
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}
