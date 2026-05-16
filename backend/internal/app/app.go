package app

import (
	"context"
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

	if err := ensureDatabaseSchema(db); err != nil {
		return err
	}

	repo := mysqlrepo.NewInterviewRepository(db)
	providerFactory := service.NewProviderFactory(a.config)
	svc := service.NewInterviewService(repo, providerFactory)
	authService := service.NewAuthService(repo, a.config.Auth)
	adminService := service.NewAdminService(repo, a.config.Auth)
	resumeParser := service.NewResumeParser(a.config.Runtime.PythonPath)

	if err := ensureSeedUser(authService); err != nil {
		return err
	}
	if err := adminService.EnsureSeedAdmin(
		context.Background(),
		a.config.Admin.SeedEmail,
		a.config.Admin.SeedPassword,
		a.config.Admin.SeedNickname,
	); err != nil {
		return err
	}

	apiHandler := handler.NewAPIHandler(svc, authService, adminService, resumeParser)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.config.Server.Port),
		Handler:           apiHandler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("api server listening on http://localhost:%d\n", a.config.Server.Port)
	fmt.Printf("default ai provider: %s\n", providerFactory.Default().Name())
	fmt.Printf("admin seed account: %s\n", a.config.Admin.SeedEmail)
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

func ensureDatabaseSchema(db *sql.DB) error {
	if db == nil {
		return nil
	}

	if _, err := db.Exec(`
		ALTER TABLE users
		MODIFY COLUMN avatar_url MEDIUMTEXT NULL
	`); err != nil {
		return fmt.Errorf("ensure users.avatar_url column: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS admins (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			email VARCHAR(120) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			nickname VARCHAR(60) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return fmt.Errorf("ensure admins table: %w", err)
	}

	return nil
}

func ensureSeedUser(authService *service.AuthService) error {
	if authService == nil {
		return nil
	}

	_, err := authService.Register(
		context.Background(),
		service.RegisterRequest{
			Email:    "demo@ai-interview.local",
			Password: "123456",
			Nickname: "演示用户",
		},
	)
	if err != nil && err.Error() != "email already exists" {
		return err
	}

	return nil
}
