package napcat_go_sdk

import (
	"time"

	"snail.local/snailllllll/utils"
)

func Rebuild_title(id string) {
	lockKey := "rebuild_title_" + id
	// 使用GetMessageViewTitle方法获取title
	title, err := GetMessageViewTitle(id)
	if err != nil {
		return
	}

	if utils.LockExists(lockKey) {
		RebuildTitlelock(&title, &utils.Config.InformGroup)

		return
	} else {
		utils.SetLock(lockKey, 300*time.Second)
		defer utils.DeleteLock(lockKey)
		RebuildTitleInform(&title, &utils.Config.InformGroup)
		ProcessForwardViewsToDB(id)
	}
}
