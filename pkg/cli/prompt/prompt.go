// Package prompt 封装基于 survey/v2 的交互式终端提示工具。
//
// 支持三种提示类型：
//   - Confirm  — 是/否确认
//   - Select   — 单选列表
//   - Input    — 文本输入
//
// 在非 TTY 环境（CI / 管道重定向）时，所有函数均优雅降级：
//   - Confirm 返回 false（不执行），调用者须配合 --yes flag 绕过
//   - Select  返回选项列表第一项
//   - Input   返回 defaultVal
//
// 典型用法：
//
//	ok, err := prompt.Confirm("确认清空数据库？")
//	if err != nil || !ok {
//	    return errors.New("已取消")
//	}
package prompt

import (
	"errors"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/term"
)

// ErrAborted 表示用户在交互中主动取消操作（Ctrl+C）。
var ErrAborted = errors.New("操作已被用户取消")

// isTTY 检测当前标准输入是否为终端（TTY）。
// 在 CI / 管道 / 非交互环境下返回 false。
func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Confirm 显示是/否确认提示，返回用户选择。
//
// 当处于非 TTY 环境时，直接返回 (false, nil)，
// 调用方应结合 --yes flag 在非交互场景下绕过此步骤。
func Confirm(question string) (bool, error) {
	if !isTTY() {
		return false, nil
	}

	var answer bool
	prompt := &survey.Confirm{
		Message: question,
		Default: false,
	}

	if err := survey.AskOne(prompt, &answer); err != nil {
		// survey 在用户按 Ctrl+C 时返回 io.EOF 或 interrupt 相关错误
		return false, ErrAborted
	}

	return answer, nil
}

// Select 显示单选列表提示，返回用户选中的选项字符串。
//
// 当处于非 TTY 环境时，返回 opts[0]（第一项）作为默认值。
// opts 不能为空，否则 panic。
func Select(question string, opts []string) (string, error) {
	if len(opts) == 0 {
		panic("prompt.Select: opts must not be empty")
	}

	if !isTTY() {
		return opts[0], nil
	}

	var answer string
	prompt := &survey.Select{
		Message: question,
		Options: opts,
	}

	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", ErrAborted
	}

	return answer, nil
}

// Input 显示文本输入提示，返回用户输入的字符串。
//
// 当处于非 TTY 环境时，直接返回 defaultVal。
func Input(question, defaultVal string) (string, error) {
	if !isTTY() {
		return defaultVal, nil
	}

	var answer string
	prompt := &survey.Input{
		Message: question,
		Default: defaultVal,
	}

	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", ErrAborted
	}

	return answer, nil
}
