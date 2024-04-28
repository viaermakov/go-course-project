//go:build integration

package tests

import (
	"github.com/stretchr/testify/suite"
	"route256.ozon.ru/project/cart/tests/suits"
	"testing"
)

func TestIntegrationSmokeSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(suits.CartServiceSuit))
}
