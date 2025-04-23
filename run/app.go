package run

import (
	"context"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"net/http"
	"os"
	"swagger_petstore/internal/handler"
	"swagger_petstore/internal/repository"
	"swagger_petstore/internal/service"
	"swagger_petstore/middleware"
	"swagger_petstore/petstore"
	"swagger_petstore/responder"
	"swagger_petstore/server"

	jsoniter "github.com/json-iterator/go"

	"github.com/jmoiron/sqlx"
	"github.com/ptflp/godecoder"

	"golang.org/x/sync/errgroup"
)

const (
	NoError = iota
	InternalError
	GeneralError
)

// Application - интерфейс приложения
type Application interface {
	Runner
	Bootstraper
}

// Runner - интерфейс запуска приложения
type Runner interface {
	Run() int
}

// Bootstraper - интерфейс инициализации приложения
type Bootstraper interface {
	Bootstrap(options ...interface{}) Runner
}

// App - структура приложения
type App struct {
	logger *zap.Logger
	db     *sqlx.DB
	srv    *server.Server
	Sig    chan os.Signal
}

// NewApp - конструктор приложения
func NewApp(db *sqlx.DB, logger *zap.Logger) *App {
	return &App{db: db, logger: logger, Sig: make(chan os.Signal, 1)}
}

// Run - запуск приложения
func (a *App) Run() int {
	ctx, cancel := context.WithCancel(context.Background())

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		sigInt := <-a.Sig
		a.logger.Info("signal interrupt recieved", zap.Stringer("os_signal", sigInt))
		cancel()
		return nil
	})

	errGroup.Go(func() error {
		err := a.srv.Serve(ctx)
		if err != nil && err != http.ErrServerClosed {
			a.logger.Error("app: server error", zap.Error(err))
			return err
		}
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return GeneralError
	}

	return NoError
}

func (a *App) Bootstrap(options ...interface{}) Runner {
	decoder := godecoder.NewDecoder(jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		DisallowUnknownFields:  true,
	})
	respond := responder.NewResponder(decoder, a.logger)

	r := chi.NewRouter()
	token := middleware.NewTokenManager(a.db)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json")))
	middlewares := []petstore.MiddlewareFunc{
		token.TokenMiddleware,
		token.BlacklistMiddleware,
	}

	rep := repository.NewRepository(a.db)
	serv := service.NewService(rep)
	optionsServer := petstore.ChiServerOptions{
		BaseRouter:  r,
		Middlewares: middlewares,
	}
	h := petstore.HandlerWithOptions(handler.NewAPI(respond, serv), optionsServer)
	a.srv = server.NewServer(h)

	return a
}
