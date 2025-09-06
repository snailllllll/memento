package napcat_go_sdk

import (
	"fmt"
	"time"

	"snail.local/snailllllll/utils"
)

func Rebuild_title(id string, username string) error {
	lockKey := "rebuild_title_" + id

	// 检查锁是否存在
	if utils.LockExists(lockKey) {
		return fmt.Errorf("对话 %s 的重命名任务已在进行中，请耐心等待", id)
	}

	// 获取锁
	utils.SetLock(lockKey, 300*time.Second)

	// 使用goroutine异步执行重命名
	go func() {
		// 确保在goroutine结束时释放锁
		defer utils.DeleteLock(lockKey)

		// 使用GetMessageViewTitle方法获取title
		title, err := GetMessageViewTitle(id)
		if err != nil {
			// 记录错误但不影响锁的释放
			fmt.Printf("获取对话标题失败: %v\n", err)
			return
		}

		// 执行重命名操作
		RebuildTitleInform(&title, &utils.Config.InformGroup, &username)
		ProcessForwardViewsToDB(id)
	}()

	// 立即返回成功发起消息
	return nil
}
