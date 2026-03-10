package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
)

func TestPackageInfoToService(t *testing.T) {
	t.Parallel()
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, PackageInfoToService(nil))
	})
	t.Run("maps weight", func(t *testing.T) {
		in := &commonpb.PackageInfo{WeightKg: 2.5}
		out := PackageInfoToService(in)
		assert.Equal(t, 2.5, out.WeightKg.GetValue())
	})
}
