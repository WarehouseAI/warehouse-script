package dependencies

import (
	"os"
	"os/signal"
	"syscall"

	authAdpt "github.com/warehouse/ai-service/internal/adapter/auth"
	mailAdpt "github.com/warehouse/ai-service/internal/adapter/mail"
	randomAdpt "github.com/warehouse/ai-service/internal/adapter/random"
	timeAdpt "github.com/warehouse/ai-service/internal/adapter/time"
	"github.com/warehouse/ai-service/internal/broker"
	"github.com/warehouse/ai-service/internal/config"
	"github.com/warehouse/ai-service/internal/db"
	"github.com/warehouse/ai-service/internal/handler/http"
	"github.com/warehouse/ai-service/internal/handler/middlewares"
	"github.com/warehouse/ai-service/internal/pkg/logger"
	nodesRepo "github.com/warehouse/ai-service/internal/repository/operations/nodes"
	scriptRepo "github.com/warehouse/ai-service/internal/repository/operations/script"
	transactionsRepo "github.com/warehouse/ai-service/internal/repository/operations/transactions"
	"github.com/warehouse/ai-service/internal/server"
	nodeSvc "github.com/warehouse/ai-service/internal/service/node"
	scriptSvc "github.com/warehouse/ai-service/internal/service/script"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Dependencies interface {
		Close()
		Cfg() *config.Config
		Internal() dependencies
		WaitForInterrupr()

		AppServer() server.Server
	}

	dependencies struct {
		cfg                     *config.Config
		log                     logger.Logger
		warehouseRequestHandler http.WarehouseRequestHandler
		handlerMiddleware       middlewares.Middleware

		psqlClient   *db.PostgresClient
		rabbitClient *broker.RabbitClient

		scriptHandler http.Handler
		nodeHandler   http.Handler

		scriptService scriptSvc.Service
		nodeService   nodeSvc.Service

		pgxTransactionRepo transactionsRepo.Repository
		scriptRepo         scriptRepo.Repository
		nodesRepo          nodesRepo.Repository

		timeAdapter   timeAdpt.Adapter
		randomAdapter randomAdpt.Adapter
		authAdapter   authAdpt.Adapter
		mailAdapter   mailAdpt.Adapter

		appServer server.Server

		shutdownChannel chan os.Signal
		closeCallbacks  []func()
	}
)

func NewDependencies(cfgPath string) (Dependencies, error) {
	cfg, err := config.NewConfig(cfgPath)
	if err != nil && err.Error() == "Config File \"config\" Not Found in \"[]\"" {
		cfg, err = config.NewConfig("./configs/local")
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "1"
	encoderCfg.TimeKey = "t"

	z := zap.New(
		&logger.WarehouseZapCore{
			Core: zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderCfg),
				zapcore.Lock(os.Stdout),
				zap.NewAtomicLevel(),
			),
		},
		zap.AddCaller(),
	)

	return &dependencies{
		cfg:             cfg,
		log:             logger.NewLogger(z),
		shutdownChannel: make(chan os.Signal),
	}, nil
}

func (d *dependencies) Close() {
	for i := len(d.closeCallbacks) - 1; i >= 0; i-- {
		d.closeCallbacks[i]()
	}
	d.log.Zap().Sync()
}

func (d *dependencies) Internal() dependencies {
	return *d
}

func (d *dependencies) Cfg() *config.Config {
	return d.cfg
}

func (d *dependencies) WarehouseJsonRequestHandler() http.WarehouseRequestHandler {
	if d.warehouseRequestHandler == nil {
		d.warehouseRequestHandler = http.NewWarehouseJsonRequestHandler(d.log, d.cfg.Timeouts.AccCookie)
	}

	return d.warehouseRequestHandler
}

func (d *dependencies) AppServer() server.Server {
	if d.appServer == nil {
		var err error
		msg := "initialize app server"
		if d.appServer, err = server.NewAppServer(
			d.log,
			d.cfg.Server,
			d.HandlerMiddleware(),
			d.ScriptHandler(),
			d.NodeHandler(),
		); err != nil {
			d.log.Zap().Panic(msg, zap.Error(err))
		}

		d.closeCallbacks = append(d.closeCallbacks, func() {
			msg := "shutting down app server"
			if err := d.appServer.Stop(); err != nil {
				d.log.Zap().Warn(msg, zap.Error(err))
				return
			}
			d.log.Zap().Info(msg)
		})
	}
	return d.appServer
}

func (d *dependencies) WaitForInterrupr() {
	signal.Notify(d.shutdownChannel, syscall.SIGINT, syscall.SIGTERM)
	d.log.Zap().Info("Wait for receive interrupt signal")
	<-d.shutdownChannel // ждем когда сигнал запишется в канал и сразу убираем его, значит, что сигнал получен
	d.log.Zap().Info("Receive interrupt signal")
}
