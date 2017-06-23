package cmd

import (
	. "forctl/io"
	"skywalker/rpc"
)

/* 处理start命令的结果 */
func processReloadResponse(v interface{}) error {
	rep := v.(*rpc.ReloadResponse)
	for _, d := range rep.GetUnchanged() {
		Print("%s - UNCHANGED\n", d)
	}
	for _, d := range rep.GetAdded() {
		Print("%s - ADDED\n", d)
	}
	for _, d := range rep.GetDeleted() {
		Print("%s - DELETED\n", d)
	}
	for _, d := range rep.GetUpdated() {
		Print("%s - UPDATED\n", d)
	}

	return nil
}
