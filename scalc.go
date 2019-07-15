package scalc

import (
	"errors"
	"reflect"
)

// Calculators опредделяет экспортируемый из модуля тип калькулятора
type Calculators struct {
	operators []operators
}

// Exec выполняет вырадение calc с набором параметров data и возвращает едиснвенное значение
// Если по завершению выполнеия вырадения кол-во значений в стеке не равно 1 - возвразается ошибка
func (calc *Calculators) Exec(data map[string]interface{}) (result interface{}, err error) {
	do := &does{make([]interface{}, 0, 16), data}
	if err = do.exec(calc.operators); err != nil {
		return
	} else if len(do.stack) != 1 {
		err = errors.New("the resulting stack size is not equal to one")
	} else {
		result = do.stack[0]
	}
	return
}

// ExecToSlice выполняет выражение calc с набором параметров data и возвращает все значения,
// находящиеся в стеке послезавершения выполнения выражения
func (calc *Calculators) ExecToSlice(data map[string]interface{}) (result []interface{}, err error) {
	do := &does{make([]interface{}, 0, 16), data}
	if err = do.exec(calc.operators); err == nil {
		result = do.stack
	}
	return
}

// does определяет исполнителя, вычисляющего выражение
type does struct {
	stack []interface{}          // стек интерпретатора выражения
	args  map[string]interface{} // набор параметров, вереданный в Calculators.Exec / Calculators.ExecToSlice
}

// operators определяет сигнатуру операций (команд) калькулятора
type operators func(*does)

// exec выполняет заданную ops последовательность операций (выражение) калькулятора
func (calc *does) exec(ops []operators) (err error) {
	defer func() {
		if temp := recover(); temp != nil {
			err = temp.(error)
		}
	}()
	for _, op := range ops {
		op(calc)
	}
	return
}

// unaryActions определяет массив унарных действий (по одной функции на каждый опустимый тип значения)
type unaryActions map[reflect.Kind]func(interface{}) interface{}

// operatorUnary является фабрикой унарных операций:
// получает на вход массив унарных действий и возвращает замыкание - операцию
func operatorUnary(action unaryActions) operators {
	return func(do *does) {
		last := len(do.stack) - 1
		do.stack[last] = action[reflect.TypeOf(do.stack[last]).Kind()](do.stack[last])
	}
}

// two определяет ключ бинарного действия: комбинацю типов двух значений
type two [2]reflect.Kind

// unaryActions определяет массив бинарных действий (по одной функции на каждую опустимую комбинацию типов двух значения)
type binaryActions map[two]func(interface{}, interface{}) interface{}

// operatorUnary является фабрикой бинарных операций:
// получает на вход массив бинарных действий и возвращает замыкание - операцию
func operatorBinary(action binaryActions) operators {
	return func(do *does) {
		last := len(do.stack) - 1
		do.stack[last-1] = action[two{
			reflect.TypeOf(do.stack[last-1]).Kind(),
			reflect.TypeOf(do.stack[last]).Kind(),
		}](do.stack[last-1], do.stack[last])
		do.stack = do.stack[:last]
	}
}

// operatorUnary является фабрикой операции, помещающей в стек значение константы:
// получает на вход значение онстанты и возвращает замыкание - операцию
func operatorConstant(value interface{}) operators {
	return func(do *does) {
		do.stack = append(do.stack, value)
	}
}

// operatorSelect реализует операцию ветвления (switch)
func operatorSelect(expressions [][]operators) operators {
	return func(do *does) {
		last := len(do.stack) - 1
		code := do.stack[last].(int64)
		do.stack = do.stack[:last]
		if err := do.exec(expressions[code]); err != nil {
			panic(err)
		}
	}
}
