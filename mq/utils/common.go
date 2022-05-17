/**
 * @Author: derek
 * @Description:
 * @File: common.go
 * @Version: 1.0.0
 * @Date: 2022/5/13 20:09
 */

package utils

import "strings"

// IsContainStrArr
// @Description: 字符串是否包含数组字符串中的某一项
// @param strArr
// @param str
// @return bool
func IsContainStrArr(strArr []string, str string) bool {
	for _, v := range strArr {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}
