package scalc

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// actions опередеяет набор операций, выполняемых калькулятором
var actions = map[string]operators{
	// Работа со стеком
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

	// Работа с параметрами
	"@": func(do *does) { // Запись в стек значения параметра с заданным именем
		last := len(do.stack) - 1
		do.stack[last] = getArgument(do.args[do.stack[last].(string)])
	},

	// Преобразование типов
	"int": operatorUnary(unaryActions{ // Преобразование значения в целое число
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
	"float": operatorUnary(unaryActions{ // Преобразование значения в вещественное число
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
	"string": operatorUnary(unaryActions{ // Преобразование значения в строку
		reflect.Int64:   func(val interface{}) interface{} { return strconv.FormatInt(val.(int64), 10) },
		reflect.Float64: func(val interface{}) interface{} { return strconv.FormatFloat(val.(float64), 'g', -1, 64) },
		reflect.String:  func(val interface{}) interface{} { return val },
	}),

	// Унарные операции: число -> число
	"--": operatorUnary(unaryActions{ // Инверсия знака числа
		reflect.Int64:   func(val interface{}) interface{} { return -val.(int64) },
		reflect.Float64: func(val interface{}) interface{} { return -val.(float64) },
	}),
	"abs": operatorUnary(unaryActions{ // Модуль числа
		reflect.Int64: func(val interface{}) interface{} {
			if temp := val.(int64); temp < 0 {
				return -temp
			}
			return val
		},
		reflect.Float64: func(val interface{}) interface{} { return math.Abs(val.(float64)) },
	}),

	// Унарные операции: число -> целое
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

	// Унарные операции: целое -> целое
	"~": operatorUnary(unaryActions{ // Инверсия битов целого числа
		reflect.Int64: func(val interface{}) interface{} { return ^val.(int64) },
	}),
	"!": operatorUnary(unaryActions{ // Логическое NOT
		reflect.Int64: func(val interface{}) interface{} {
			return convertBool(val == int64(0))
		},
	}),

	// Унарные операции: вещественное -> вещественное
	"sqrt": operatorUnary(unaryActions{ // Квадратный корень
		reflect.Float64: func(val interface{}) interface{} { return math.Sqrt(val.(float64)) },
	}),
	"ln": operatorUnary(unaryActions{ // Квадратный корень
		reflect.Float64: func(val interface{}) interface{} { return math.Log(val.(float64)) },
	}),
	"exp": operatorUnary(unaryActions{ // Квадратный корень
		reflect.Float64: func(val interface{}) interface{} { return math.Exp(val.(float64)) },
	}),
	"floor": operatorUnary(unaryActions{ // Округление вниз
		reflect.Float64: func(val interface{}) interface{} { return math.Floor(val.(float64)) },
	}),
	"ceil": operatorUnary(unaryActions{ // Округление вверх
		reflect.Float64: func(val interface{}) interface{} { return math.Ceil(val.(float64)) },
	}),
	"round": operatorUnary(unaryActions{ // Округление к ближайшёму
		reflect.Float64: func(val interface{}) interface{} { return math.Round(val.(float64)) },
	}),
	"trunc": operatorUnary(unaryActions{ // Округление к ближайшёму
		reflect.Float64: func(val interface{}) interface{} { return math.Trunc(val.(float64)) },
	}),
	"frac": operatorUnary(unaryActions{ // Округление к ближайшёму
		reflect.Float64: func(val interface{}) (res interface{}) {
			_, res = math.Modf(val.(float64))
			return
		},
	}),

	// Унарные операции: вещественное -> целое
	"isNaN": operatorUnary(unaryActions{ // Проверка на NaN
		reflect.Float64: func(val interface{}) interface{} { return convertBool(math.IsNaN(val.(float64))) },
	}),
	"isInf": operatorUnary(unaryActions{ // Проверка на Inf
		reflect.Float64: func(val interface{}) interface{} { return convertBool(math.IsInf(val.(float64), 0)) },
	}),

	// Унарные операции: строка -> строка
	"trim": operatorUnary(unaryActions{ // Удаление краних пробельных символов
		reflect.String: func(val interface{}) interface{} { return strings.TrimSpace(val.(string)) },
	}),
	"upper": operatorUnary(unaryActions{ // Преобразование в верхний регистр
		reflect.String: func(val interface{}) interface{} { return strings.ToUpper(val.(string)) },
	}),
	"lower": operatorUnary(unaryActions{ // Преобразование в нижний регистр
		reflect.String: func(val interface{}) interface{} { return strings.ToLower(val.(string)) },
	}),

	// Унарные операции: строка -> целое
	"len": operatorUnary(unaryActions{ // Длина строки
		reflect.String: func(val interface{}) interface{} { return int64(len(val.(string))) },
	}),

	// Унарные операции: знечение -> целое
	"isEmpty": operatorUnary(unaryActions{ // Проверка на пустое значение
		reflect.Int64:   func(val interface{}) interface{} { return convertBool(val.(int64) == 0) },
		reflect.Float64: func(val interface{}) interface{} { return convertBool(val.(float64) == 0.0) },
		reflect.String:  func(val interface{}) interface{} { return convertBool(val.(string) == "") },
	}),

	// Бинарные операции: значение, значение -> значение
	"+": operatorBinary(binaryActions{ // Сложение чисел / конкатенация строк
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) + v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) + v2.(float64) },
		two{reflect.String, reflect.String}:   func(v1, v2 interface{}) interface{} { return v1.(string) + v2.(string) },
	}),
	"min": operatorBinary(binaryActions{ // Не равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			if v1.(int64) < v2.(int64) {
				return v1
			}
			return v2
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return math.Min(v1.(float64), v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			if v1.(string) < v2.(string) {
				return v1
			}
			return v2
		},
	}),
	"max": operatorBinary(binaryActions{ // Не равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			if v1.(int64) > v2.(int64) {
				return v1
			}
			return v2
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return math.Max(v1.(float64), v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			if v1.(string) > v2.(string) {
				return v1
			}
			return v2
		},
	}),

	// Бинарные операции: число, число -> число
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

	// Бинарные операции: вещественное, вещественное -> вещественное
	"**": operatorBinary(binaryActions{ // Возведение в степень
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return math.Pow(v1.(float64), v2.(float64)) },
	}),

	// Бинарные операции: целое, целое -> целое
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

	// Бинарные операции: значение, значение -> целое
	"=": operatorBinary(binaryActions{ // Равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(int64) == v2.(int64))
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(float64) == v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(string) == v2.(string))
		},
	}),
	"#": operatorBinary(binaryActions{ // Не равно
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(int64) != v2.(int64))
		},
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(float64) != v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return convertBool(v1.(string) != v2.(string))
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

	// Бинарные операции: строка, строка -> целое
	"index": operatorBinary(binaryActions{ // Поиск позиции первого вхождения подстроки
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return int64(strings.Index(v1.(string), v2.(string)))
		},
	}),
	"indexLast": operatorBinary(binaryActions{ // Поиск позиции последнего вхождения подстроки
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return int64(strings.LastIndex(v1.(string), v2.(string)))
		},
	}),
	"timeParse": operatorBinary(binaryActions{ // Преобразование записи даты/времени в числовую метку времени
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			if tm, err := time.Parse(v1.(string), v2.(string)); err != nil {
				panic(err)
			} else {
				return tm.Unix()
			}
		},
	}),
	"regexMatch": operatorBinary(binaryActions{ // Проверка на соотвествие шаблону
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			if result, err := regexp.MatchString(v1.(string), v2.(string)); err != nil {
				panic(err)
			} else {
				return convertBool(result)
			}
		},
	}),

	// Бинарные операции: строка, целое -> строка
	"left": operatorBinary(binaryActions{ // Левая часть строки
		two{reflect.String, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return v1.(string)[:v2.(int64)]
		},
	}),
	"right": operatorBinary(binaryActions{ // Правая часть строки
		two{reflect.String, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			temp := v1.(string)
			return temp[int64(len(temp))-v2.(int64):]
		},
	}),
	"timeFormat": operatorBinary(binaryActions{ // Преобразование числовой метки времени в запись даты/времени
		two{reflect.String, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return time.Unix(v2.(int64), 0).Format(v1.(string))
		},
	}),

	// Бинарные операции: срока, значение -> строка
	"format": operatorBinary(binaryActions{ // Форматирование значения
		two{reflect.String, reflect.Int64}: func(v1, v2 interface{}) interface{} {
			return fmt.Sprintf("%"+v1.(string), v2.(int64))
		},
		two{reflect.String, reflect.Float64}: func(v1, v2 interface{}) interface{} {
			return fmt.Sprintf("%"+v1.(string), v2.(float64))
		},
		two{reflect.String, reflect.String}: func(v1, v2 interface{}) interface{} {
			return fmt.Sprintf("%"+v1.(string), v2.(string))
		},
	}),

	// Тернарные операции строка, строка, строка -> строка
	"replace": operatorTernary(ternaryActions{ // Замена подстроки
		three{reflect.String, reflect.String, reflect.String}: func(v1, v2, v3 interface{}) interface{} {
			return strings.ReplaceAll(v3.(string), v1.(string), v2.(string))
		},
	}),
	"regexReplace": operatorTernary(ternaryActions{ // Замена регулярного выражения
		three{reflect.String, reflect.String, reflect.String}: func(v1, v2, v3 interface{}) interface{} {
			if regex, err := regexp.Compile(v1.(string)); err != nil {
				panic(err)
			} else {
				return regex.ReplaceAllString(v3.(string), v2.(string))
			}
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
