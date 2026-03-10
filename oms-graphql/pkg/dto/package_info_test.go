package dto //nolint:testpackage // testing exported API only

import (
	"testing"

	"github.com/stretchr/testify/assert"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
)

func TestPackageInfoToService(t *testing.T) {
	t.Parallel()
	t.Run("nil returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, PackageInfoToService(nil))
	})
	t.Run("maps weight", func(t *testing.T) {
		t.Parallel()

		in := &commonpb.PackageInfo{WeightKg: 2.5}
		out := PackageInfoToService(in)
		assert.InEpsilon(t, 2.5, out.GetWeightKg().GetValue(), 1e-9)
	})
}
