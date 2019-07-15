package easy

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// actions опередеяет набор операций, выполняемых калькулятором
var actions = map[string]operators{
	"drop": func(do *does) { // Удаление значения из вершины стека
		do.stack = do.stack[:len(do.stack)-1]
	},
	"dup": func(do *does) { // Дублирование вершины стека
		do.stack = append(do.stack, do.stack[len(do.stack)-1])
	},
	"swap": func(do *does) { // Обмен двух значений в врешине стека
		last := len(do.stack) - 1
		temp := do.stack[last]
		do.stack[last] = do.stack[last-1]
		do.stack[last-1] = temp
	},
	"over": func(do *does) { // Запись в стек второго от вершины значния
		do.stack = append(do.stack, do.stack[len(do.stack)-2])
	},

	"@": func(do *does) { // Запись в стек значения параметра с заданным именем
		last := len(do.stack) - 1
		do.stack[last] = getArgument(do.args[do.stack[last].(string)])
	},

	"int": operatorUnary(unaryActions{ // Преобразование значения в вершине стека в целое число
		reflect.Int64:   func(val interface{}) interface{} { return val },
		reflect.Float64: func(val interface{}) interface{} { return int64(val.(float64)) },
		reflect.String: func(val interface{}) interface{} {
			if result, err := strconv.ParseInt(val.(string), 10, 64); err != nil {
				panic(err)
			} else {
				return result
			}
		},
	}),
	"float": operatorUnary(unaryActions{ // Преобразование значения в вершине стека в вещественное число
		reflect.Int64:   func(val interface{}) interface{} { return float64(val.(int64)) },
		reflect.Float64: func(val interface{}) interface{} { return val },
		reflect.String: func(val interface{}) interface{} {
			if result, err := strconv.ParseFloat(strings.ReplaceAll(val.(string), ",", "."), 64); err != nil {
				panic(err)
			} else {
				return result
			}
		},
	}),
	"string": operatorUnary(unaryActions{ // Преобразование значения в вершине стека в строку
		reflect.Int64:   func(val interface{}) interface{} { return strconv.FormatInt(val.(int64), 10) },
		reflect.Float64: func(val interface{}) interface{} { return strconv.FormatFloat(val.(float64), 'g', -1, 64) },
		reflect.String:  func(val interface{}) interface{} { return val },
	}),

	"--": operatorUnary(unaryActions{ // Инверсия знака числа
		reflect.Int64:   func(val interface{}) interface{} { return -val.(int64) },
		reflect.Float64: func(val interface{}) interface{} { return -val.(float64) },
	}),
	"~": operatorUnary(unaryActions{ // Инверсия битов целого числа
		reflect.Int64: func(val interface{}) interface{} { return ^val.(int64) },
	}),
	"!": operatorUnary(unaryActions{ // Логическое NOT
		reflect.Int64: func(val interface{}) interface{} {
			return convertBool(val == int64(0))
		},
	}),
	"abs": operatorUnary(unaryActions{ // Молуль числа
		reflect.Int64: func(val interface{}) interface{} {
			if temp := val.(int64); temp < 0 {
				return -temp
			}
			return val
		},
		reflect.Float64: func(val interface{}) interface{} { return math.Abs(val.(float64)) },
	}),
	"sign": operatorUnary(unaryActions{ // Знак числа
		reflect.Int64: func(val interface{}) interface{} {
			if temp := val.(int64); temp < 0 {
				return int64(-1)
			} else {
				return convertBool(temp > 0)
			}
		},
		reflect.Float64: func(val interface{}) interface{} {
			if temp := val.(float64); temp < 0.0 {
				return int64(-1)
			} else {
				return convertBool(temp > 0.0)
			}
		},
	}),
	"len": operatorUnary(unaryActions{ // Длина строки
		reflect.String: func(val interface{}) interface{} { return int64(len(val.(string))) },
	}),
	"sqrt": operatorUnary(unaryActions{ // Квадратный корень
		reflect.Float64: func(val interface{}) interface{} { return math.Sqrt(val.(float64)) },
	}),
	"isNaN": operatorUnary(unaryActions{ // Проверка на NaN
		reflect.Float64: func(val interface{}) interface{} { return convertBool(math.IsNaN(val.(float64))) },
	}),
	"isInf": operatorUnary(unaryActions{ // Проверка на Inf
		reflect.Float64: func(val interface{}) interface{} { return convertBool(math.IsInf(val.(float64), 0)) },
	}),

	"+": operatorBinary(binaryActions{ // Сложение чисел / конкатенация строк
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) + v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) + v2.(float64) },
		two{reflect.String, reflect.String}:   func(v1, v2 interface{}) interface{} { return v1.(string) + v2.(string) },
	}),
	"-": operatorBinary(binaryActions{ // Вычитание
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) - v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) - v2.(float64) },
	}),
	"*": operatorBinary(binaryActions{ // Умножение
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) * v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) * v2.(float64) },
	}),
	"/": operatorBinary(binaryActions{ // Деление
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) / v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) / v2.(float64) },
	}),
	"%": operatorBinary(binaryActions{ // Остаток от деления целых чисел
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) % v2.(int64) },
	}),
	"&": operatorBinary(binaryActions{ // Битовое AND целых чисел
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) & v2.(int64) },
	}),
	"|": operatorBinary(binaryActions{ // Битовое OR целых чисел
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) | v2.(int64) },
	}),
	"^": operatorBinary(binaryActions{ // Битовое XOR целых чисел
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) ^ v2.(int64) },
	}),
	"<<": operatorBinary(binaryActions{ // Битовый сдвиг влево целого числа
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) << uint64(v2.(int64)) },
	}),
	">>": operatorBinary(binaryActions{ // Битовый сдвиг враво целого числа
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) >> uint64(v2.(int64)) },
	}),

	"=": operatorBinary(binaryActions{ // Равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1 == v2)
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1 == v2)
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1 == v2)
		},
	}),
	"!=": operatorBinary(binaryActions{ // Не равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1 != v2)
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1 != v2)
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1 != v2)
		},
	}),
	">": operatorBinary(binaryActions{ // Больше
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(int64) > v2.(int64))
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(float64) > v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(string) > v2.(string))
		},
	}),
	"<": operatorBinary(binaryActions{ // Меньше
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(int64) < v2.(int64))
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(float64) < v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(string) < v2.(string))
		},
	}),
	">=": operatorBinary(binaryActions{ // Больше или равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(int64) >= v2.(int64))
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(float64) >= v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(string) >= v2.(string))
		},
	}),
	"<=": operatorBinary(binaryActions{ // Меньше или равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(int64) <= v2.(int64))
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(float64) <= v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(string) <= v2.(string))
		},
	}),
}

// convertBool преобразует значение типа bool в int64: false -> 0, true -> 1
func convertBool(flg bool) int64 {
	if flg {
		return int64(1)
	} else {
		return int64(0)
	}
}

// getArgument получает значение параметра выражения по его имени - с преобразованием его значения в допустимый тип
func getArgument(value interface{}) interface{} {
	switch value := reflect.Indirect(reflect.ValueOf(value)); value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int()
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.String:
		return value.String()
	default:
		panic(errors.New("argument type is not valid"))
	}
}
