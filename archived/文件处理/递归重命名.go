package 文件处理

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RenameInDirectory 递归地重命名目录及其子目录和文件中包含指定字符串的部分
// rootPath: 根目录路径
// oldStr: 要被替换的字符串
// newStr: 替换后的字符串（可以为空字符串）
func RenameInDirectory(rootPath, oldStr, newStr string) error {
	if oldStr == "" {
		return fmt.Errorf("oldStr 不能为空")
	}

	// 收集所有需要重命名的路径
	var pathsToRename []string

	// 先遍历一遍，收集所有需要重命名的路径
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查文件名或目录名是否包含要替换的字符串
		if strings.Contains(info.Name(), oldStr) {
			pathsToRename = append(pathsToRename, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %w", err)
	}

	// 从最深的路径开始重命名（反向遍历），避免父目录先被重命名导致子路径失效
	for i := len(pathsToRename) - 1; i >= 0; i-- {
		oldPath := pathsToRename[i]
		dir := filepath.Dir(oldPath)
		oldName := filepath.Base(oldPath)
		newName := strings.ReplaceAll(oldName, oldStr, newStr)
		newPath := filepath.Join(dir, newName)

		// 检查新路径是否已存在
		if _, err := os.Stat(newPath); err == nil {
			fmt.Printf("警告: 目标路径已存在，跳过: %s -> %s\n", oldPath, newPath)
			continue
		}

		// 执行重命名
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return fmt.Errorf("重命名失败 %s -> %s: %w", oldPath, newPath, err)
		}

		fmt.Printf("成功重命名: %s -> %s\n", oldPath, newPath)
	}

	return nil
}
