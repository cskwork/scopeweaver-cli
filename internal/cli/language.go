package cli

import "strings"

func errorSummary(lang, code string) string {
	messages := map[string][3]string{
		"invalid_input":         {"Invalid command input. Inspect the command schema.", "명령 입력이 올바르지 않습니다. 명령 스키마를 확인하세요.", "命令输入无效。请检查命令模式。"},
		"unknown_command":       {"Unknown command. Use help or commands.", "알 수 없는 명령입니다. help 또는 commands를 확인하세요.", "未知命令。请使用 help 或 commands。"},
		"confirmation_required": {"Explicit confirmation is required for this operation.", "이 작업에는 명시적인 확인이 필요합니다.", "此操作需要明确确认。"},
		"runtime_unavailable":   {"The private runtime could not be started or reached.", "전용 런타임을 시작하거나 연결할 수 없습니다.", "无法启动或连接专用运行环境。"},
		"runtime_error":         {"The private runtime operation failed.", "전용 런타임 작업에 실패했습니다.", "专用运行环境操作失败。"},
		"backend_error":         {"The engine rejected the operation. See the diagnostic detail.", "엔진이 작업을 거부했습니다. 진단 내용을 확인하세요.", "引擎拒绝了操作。请查看诊断详情。"},
		"request_failed":        {"The operation failed. Inspect state before retrying a write.", "작업에 실패했습니다. 변경 작업을 재시도하기 전에 상태를 확인하세요.", "操作失败。重试写入之前请检查状态。"},
		"input_error":           {"The input could not be read.", "입력을 읽을 수 없습니다.", "无法读取输入。"},
		"cancelled":             {"The operation was cancelled.", "작업이 취소되었습니다.", "操作已取消。"},
	}
	value, ok := messages[code]
	if !ok {
		value = messages["request_failed"]
	}
	switch lang {
	case "ko":
		return value[1]
	case "zh-Hans":
		return value[2]
	default:
		return value[0]
	}
}
func redactMessage(message string, value any) string {
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			if secretKey.MatchString(key) {
				if secret, ok := item.(string); ok && secret != "" {
					message = replaceSecret(message, secret)
				}
			} else {
				message = redactMessage(message, item)
			}
		}
	case []any:
		for _, item := range v {
			message = redactMessage(message, item)
		}
	}
	return message
}

func replaceSecret(message, secret string) string {
	return strings.ReplaceAll(message, secret, "[REDACTED]")
}

func redactQuery(values map[string][]string) map[string][]string {
	result := map[string][]string{}
	for key, value := range values {
		if secretKey.MatchString(key) {
			result[key] = []string{"[REDACTED]"}
		} else {
			result[key] = value
		}
	}
	return result
}
