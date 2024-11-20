package ojocronlisteners

import (
	"github.com/devzatruk/bizhubBackend/config"
)

func AddOjoCronListeners() {
	config.OjoCronService.On("discount_removed", AutoPostDiscountRemoved)
	config.OjoCronService.On("add_discount", AutoPostAddDiscount)
	config.OjoCronService.On("auction_finished", HandleAuctionFinished)
	config.OjoCronService.On("auction_removed", HandleAuctionRemoved)
	config.OjoCronService.On(config.PERMISSION_STARTED, HandlePermissionStarted)
	config.OjoCronService.On(config.PERMISSION_ENDED, HandlePermissionEnded)
}
