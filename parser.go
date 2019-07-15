package scalc

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	// splitPattern содержит шаблон разюивки исходного выражения на лексемы
	splitPattern = regexp.MustCompile("\\s+")
	// escapePattern содержит шаблон посика в лексеме "строковая константа" специальных символов
	escapePattern = regexp.MustCompile("\\\\.")
)

// New получает на вход строку, содержащую выражение, и возвращает экземпляр калькулятора, вычисляющего это выражение
func New(expr string) (*Calculators, error) {
	buffer := [][][]operators{{{}}}

	level := 0
	section := 0
	for _, lexeme := range splitPattern.Split(strings.TrimSpace(expr), -1) {
		if op, exists := actions[lexeme]; exists {
			buffer[level][section] = append(buffer[level][section], op)
		} else if lexeme == "[" {
			buffer = append(buffer, [][]operators{{}})
			level++
			section = 0
		} else if lexeme == "]" {
			if level == 0 {
				return nil, errors.New("] without [")
			}
			temp := operatorSelect(buffer[level])
			buffer = buffer[:level]
			level--
			section = len(buffer[level]) - 1
			buffer[level][section] = append(buffer[level][section], temp)
		} else if lexeme == ";" {
			if level == 0 {
				return nil, errors.New("; outside []")
			}
			buffer[level] = append(buffer[level], []operators{})
			section++
		} else if value, err := strconv.ParseInt(lexeme, 10, 64); err == nil {
			buffer[level][section] = append(buffer[level][section], operatorConstant(value))
		} else if value, err := strconv.ParseFloat(lexeme, 64); err == nil {
			buffer[level][section] = append(buffer[level][section], operatorConstant(value))
		} else if len(lexeme) > 0 && lexeme[0] == '\'' {
			buffer[level][section] = append(buffer[level][section], operatorConstant(convertString(lexeme[1:])))
		} else {
			buffer[level][section] = append(buffer[level][section], operatorConstant(convertString(lexeme)))
		}
	}
	if level > 0 {
		return nil, errors.New("[ without ]")
	}
	return &Calculators{buffer[0][0]}, nil
}

// convertString производит замену в строке специальных символов
func convertString(str string) string {
	return escapePattern.ReplaceAllStringFunc(str, func(str string) string {
		switch str[1] {
		case 's':
			return " "
		case 'n':
			return "\n"
		case 't':
			return "\t"
		case '\\':
			return "\\"
		default:
			return string(str[1])
		}
	})
}
