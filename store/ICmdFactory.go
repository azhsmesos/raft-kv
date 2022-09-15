package store

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

// ICmdFactory 操作指令工厂
type ICmdFactory interface {
	OfTag(tag int) ICmd
	Tag(cmd ICmd) int
}

type defaultCmdFactory int

const (
	putCmdTag = 1
	delCmdTag = 2
)

var (
	cmdFactory = new(defaultCmdFactory)
	log        = logrus.New()
)

func (dcf *defaultCmdFactory) OfTag(tag int) ICmd {
	switch tag {
	case putCmdTag:
		return new(PutCmd)
	case delCmdTag:
		return new(DelCmd)
	}
	panic(fmt.Sprintf("unkonwn tag: %d", tag))
}

func (dcf *defaultCmdFactory) Tag(cmd ICmd) int {
	if _, ok := cmd.(*PutCmd); ok {
		return putCmdTag
	}
	if _, ok := cmd.(*DelCmd); ok {
		return delCmdTag
	}
	panic(fmt.Sprintf("unkonwn cmd: %v", cmd))
}
