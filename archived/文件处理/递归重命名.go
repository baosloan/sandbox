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

// RenameOptions 重命名选项
type RenameOptions struct {
	OldStr        string // 要被替换的字符串
	NewStr        string // 替换后的字符串
	CaseSensitive bool   // 是否区分大小写
	DryRun        bool   // 是否只模拟运行，不实际重命名
	SkipErrors    bool   // 是否跳过错误继续执行
}

// RenameInDirectoryWithOptions 使用选项进行重命名
func RenameInDirectoryWithOptions(rootPath string, opts RenameOptions) error {
	if opts.OldStr == "" {
		return fmt.Errorf("oldStr 不能为空")
	}

	// 收集所有需要重命名的路径
	var pathsToRename []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if opts.SkipErrors {
				fmt.Printf("警告: 访问路径失败: %s, 错误: %v\n", path, err)
				return nil
			}
			return err
		}

		fileName := info.Name()
		var contains bool

		if opts.CaseSensitive {
			contains = strings.Contains(fileName, opts.OldStr)
		} else {
			contains = strings.Contains(
				strings.ToLower(fileName),
				strings.ToLower(opts.OldStr),
			)
		}

		if contains {
			pathsToRename = append(pathsToRename, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %w", err)
	}

	if len(pathsToRename) == 0 {
		fmt.Println("没有找到需要重命名的文件或目录")
		return nil
	}

	fmt.Printf("找到 %d 个需要重命名的项目\n", len(pathsToRename))

	// 从最深的路径开始重命名
	successCount := 0
	for i := len(pathsToRename) - 1; i >= 0; i-- {
		oldPath := pathsToRename[i]
		dir := filepath.Dir(oldPath)
		oldName := filepath.Base(oldPath)

		var newName string
		if opts.CaseSensitive {
			newName = strings.ReplaceAll(oldName, opts.OldStr, opts.NewStr)
		} else {
			newName = replaceAllCaseInsensitive(oldName, opts.OldStr, opts.NewStr)
		}

		newPath := filepath.Join(dir, newName)

		// 如果新旧路径相同，跳过
		if oldPath == newPath {
			continue
		}

		// 检查新路径是否已存在
		if _, err := os.Stat(newPath); err == nil {
			msg := fmt.Sprintf("警告: 目标路径已存在，跳过: %s -> %s", oldPath, newPath)
			fmt.Println(msg)
			if !opts.SkipErrors {
				return fmt.Errorf(msg)
			}
			continue
		}

		if opts.DryRun {
			fmt.Printf("[模拟] %s -> %s\n", oldPath, newPath)
			successCount++
		} else {
			err := os.Rename(oldPath, newPath)
			if err != nil {
				msg := fmt.Sprintf("重命名失败 %s -> %s: %v", oldPath, newPath, err)
				if opts.SkipErrors {
					fmt.Printf("警告: %s\n", msg)
					continue
				}
				return fmt.Errorf(msg)
			}
			fmt.Printf("✓ %s -> %s\n", oldPath, newPath)
			successCount++
		}
	}

	fmt.Printf("\n总计: 成功 %d 个\n", successCount)
	return nil
}

// replaceAllCaseInsensitive 不区分大小写的替换
func replaceAllCaseInsensitive(s, old, new string) string {
	lowerS := strings.ToLower(s)
	lowerOld := strings.ToLower(old)

	result := ""
	lastIndex := 0

	for {
		index := strings.Index(lowerS[lastIndex:], lowerOld)
		if index == -1 {
			result += s[lastIndex:]
			break
		}

		actualIndex := lastIndex + index
		result += s[lastIndex:actualIndex] + new
		lastIndex = actualIndex + len(old)
	}

	return result
}
