package split

import (
	"split/splitwise/rest"
	"split/splitwise/service"
)

var SplitWiseService service.ISplitWiseService

func Init() {
	SplitWiseService = &service.SplitWiseService{}
	rest.RegisterService("SplitWiseService", SplitWiseService)
}
