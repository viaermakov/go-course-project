package suits

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"route256.ozon.ru/project/loms/config"
	"route256.ozon.ru/project/loms/internals/infra/db"
	"route256.ozon.ru/project/loms/internals/infra/shardmanager"
	"route256.ozon.ru/project/loms/internals/repository/notifierrepo"
	"route256.ozon.ru/project/loms/internals/repository/ordersrepo"
	"route256.ozon.ru/project/loms/internals/repository/stocksrepo"
	"route256.ozon.ru/project/loms/internals/service/lomsservice"
	"route256.ozon.ru/project/loms/migrations"
	"route256.ozon.ru/project/loms/model/itemmodel"
	"route256.ozon.ru/project/loms/model/ordermodel"
	"route256.ozon.ru/project/loms/tests/testcontainer"
)

type LomsServiceSuite struct {
	suite.Suite
	dbConnStr   string
	ctx         context.Context
	service     *lomsservice.LomsService
	pgContainer *postgres.PostgresContainer
}

func (suite *LomsServiceSuite) SetupSuite() {
	appConfig := config.NewConfig()

	ctx := context.Background()
	suite.ctx = ctx

	pgContainer, err := testcontainer.CreatePgContainer(ctx, appConfig)

	if err != nil {
		suite.Fail("Failed to create Postgres container: " + err.Error())
	}

	suite.pgContainer = pgContainer
	dbConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")

	if err != nil {
		suite.Fail("Failed to create connection string: " + err.Error())
	}

	suite.dbConnStr = dbConnStr

	dbPool, err := db.NewPool(ctx, []string{dbConnStr}, []string{})
	stocksStorage := stocksrepo.NewRepo()
	ordersStorage := ordersrepo.NewRepo()
	notifierStorage := notifierrepo.NewRepo()
	shardManager := shardmanager.New(shardmanager.GetShardFn(2), []db.Pool{dbPool, dbPool})

	suite.service = lomsservice.NewLomsService(shardManager, dbPool, stocksStorage, ordersStorage, notifierStorage)
}

func (suite *LomsServiceSuite) SetupTest() {
	migrations.ApplyMigrations(suite.dbConnStr, "test")
}

func (suite *LomsServiceSuite) TearDownTest() {
	migrations.RollbackMigrations(suite.dbConnStr, "test")
}

func (suite *LomsServiceSuite) TearDownSuit() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		suite.Fail("Failed to terminate pg container: " + err.Error())
	}
}

func (suite *LomsServiceSuite) TestCreateOrderFlowIntegration() {
	userId := int64(1)
	items := []itemmodel.Item{{SkuId: 1002, Count: 1}}

	orderId, err := suite.service.CreateOrder(suite.ctx, userId, items)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(orderId)

	stocks, err := suite.service.GetAvailableStocks(suite.ctx, int64(items[0].SkuId))
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(179), stocks)

	info, err := suite.service.GetOrder(suite.ctx, orderId)
	suite.Require().NoError(err)
	suite.Require().Equal(userId, info.User)
	suite.Require().ElementsMatch(items, info.Items)
	suite.Require().Equal(ordermodel.StatusAwaiting, info.Status)
}

func (suite *LomsServiceSuite) TestPayOrderFlowIntegration() {
	userId := int64(1)
	items := []itemmodel.Item{{SkuId: 1002, Count: 1}}

	orderId, err := suite.service.CreateOrder(suite.ctx, userId, items)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(orderId)

	err = suite.service.PayOrder(suite.ctx, orderId)
	suite.Require().NoError(err)

	info, err := suite.service.GetOrder(suite.ctx, orderId)
	suite.Require().NoError(err)
	suite.Require().Equal(ordermodel.StatusPaid, info.Status)
}

func (suite *LomsServiceSuite) TestCancelOrderFlowIntegration() {
	userId := int64(1)
	items := []itemmodel.Item{{SkuId: 1002, Count: 1}}

	orderId, err := suite.service.CreateOrder(suite.ctx, userId, items)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(orderId)

	err = suite.service.CancelOrder(suite.ctx, orderId)
	suite.Require().NoError(err)

	stocks, err := suite.service.GetAvailableStocks(suite.ctx, int64(items[0].SkuId))
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(180), stocks)

	info, err := suite.service.GetOrder(suite.ctx, orderId)
	suite.Require().NoError(err)
	suite.Require().Equal(ordermodel.StatusCanceled, info.Status)
}

func (suite *LomsServiceSuite) TestOrderNotFoundFlowIntegration() {
	_, err := suite.service.GetOrder(suite.ctx, 1)
	suite.Require().ErrorIs(err, ordersrepo.ErrOrderNotFound)
}
