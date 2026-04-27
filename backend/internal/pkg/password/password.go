package password

import (
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinLength 密码最小长度
	MinLength = 8
	// MaxLength 密码最大长度
	MaxLength = 128
	// DefaultCost bcrypt默认成本
	DefaultCost = 12
)

// Hash 对密码进行哈希
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify 验证密码
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Strength 密码强度
type Strength int

const (
	StrengthWeak   Strength = 1
	StrengthMedium Strength = 2
	StrengthStrong Strength = 3
)

// ValidationResult 密码验证结果
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Strength Strength `json:"strength"`
	Errors   []string `json:"errors,omitempty"`
}

// Validate 验证密码强度
func Validate(password string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: make([]string, 0),
	}

	// 检查长度
	if len(password) < MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, "密码长度至少为8位")
	}
	if len(password) > MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, "密码长度不能超过128位")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 计算强度
	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}
	if len(password) >= 12 {
		score++
	}

	switch {
	case score >= 4:
		result.Strength = StrengthStrong
	case score >= 3:
		result.Strength = StrengthMedium
	default:
		result.Strength = StrengthWeak
	}

	// 至少需要包含数字和字母
	if !hasDigit {
		result.Valid = false
		result.Errors = append(result.Errors, "密码必须包含数字")
	}
	if !hasUpper && !hasLower {
		result.Valid = false
		result.Errors = append(result.Errors, "密码必须包含字母")
	}

	// 检查常见弱密码模式
	weakPatterns := []string{
		`^123456`,
		`^password`,
		`^qwerty`,
		`^admin`,
		`(.)\1{3,}`, // 4个以上重复字符
	}

	for _, pattern := range weakPatterns {
		matched, _ := regexp.MatchString(pattern, password)
		if matched {
			result.Valid = false
			result.Errors = append(result.Errors, "密码过于简单")
			break
		}
	}

	return result
}

// IsStrong 判断密码是否足够强
func IsStrong(password string) bool {
	result := Validate(password)
	return result.Valid && result.Strength >= StrengthMedium
}
