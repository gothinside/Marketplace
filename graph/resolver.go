package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
import (
	"hw11_shopql/pkg/cart"
	"hw11_shopql/pkg/catalog"
	"hw11_shopql/pkg/item"
	"hw11_shopql/pkg/order"
	"hw11_shopql/pkg/role"
	"hw11_shopql/pkg/seller"
)

type Resolver struct {
	RoleRepo    role.RoleRepoI
	CatalogRepo catalog.CataloRepoInrerface
	ItemRepo    item.ItemRepoInterface
	SellerRepo  seller.SellerRepoInterface
	CartRepo    cart.CartRepoInterface
	OrderRepo   order.OrderRepoInterface
}
