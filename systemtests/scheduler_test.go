package systemtests

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/contiv/contivmodel/client"
	. "gopkg.in/check.v1"
)

type systemTestScheduler interface {
	runContainers(spec *jobSpec)
}
