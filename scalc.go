package scalc

import (
	"errors"
	"math"
	"reflect"
	"strconv"
)

var errorTable = []error{
	errors.New("stack is empty"),
	errors.New("bad operand type"),
	errors.New("the resulting stack size is not equal to one"),
	errors.New("selector is not integer"),
	errors.New("selector is not in range"),
}

type does struct {
	stack []interface{}
	args  map[string]interface{}
}

type operator func(*does) error

func (calc *does) exec(ops []operator) (err error) {
	defer func() {
		if temp := recover(); temp != nil {
			err = temp.(error)
		}
	}()
	for _, op := range ops {
		if err = op(calc); err != nil {
			return
		}
	}
	return
}

type Calc struct {
	operators []operator
}

var parser = map[string]operator{
	"drop": operatorDrop,
	"dup":  operatorDup,
	"swap": operatorSwap,
	"over": operatorOver,
	"@":    operatorArgument,
	"int": operatorUnary(unaryAction{
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
	"float": operatorUnary(unaryAction{
		reflect.Int64:   func(val interface{}) interface{} { return float64(val.(int64)) },
		reflect.Float64: func(val interface{}) interface{} { return val },
		reflect.String: func(val interface{}) interface{} {
			if result, err := strconv.ParseFloat(val.(string), 64); err != nil {
				panic(err)
			} else {
				return result
			}
		},
	}),
	"string": operatorUnary(unaryAction{
		reflect.Int64:   func(val interface{}) interface{} { return strconv.FormatInt(val.(int64), 10) },
		reflect.Float64: func(val interface{}) interface{} { return strconv.FormatFloat(val.(float64), 'g', -1, 64) },
		reflect.String:  func(val interface{}) interface{} { return val },
	}),
	"~": operatorUnary(unaryAction{
		reflect.Int64:   func(val interface{}) interface{} { return -val.(int64) },
		reflect.Float64: func(val interface{}) interface{} { return -val.(float64) },
	}),
	"!": operatorUnary(unaryAction{
		reflect.Int64: func(val interface{}) interface{} { return ^val.(int64) },
	}),
	"abs": operatorUnary(unaryAction{
		reflect.Int64: func(val interface{}) interface{} {
			if temp := val.(int64); temp < 0 {
				return -temp
			}
			return val
		},
		reflect.Float64: func(val interface{}) interface{} { return math.Abs(val.(float64)) },
	}),
	"+": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) + v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) + v2.(float64) },
		two{reflect.String, reflect.String}:   func(v1, v2 interface{}) interface{} { return v1.(string) + v2.(string) },
	}),
	"-": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) - v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) - v2.(float64) },
	}),
	"*": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) * v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) * v2.(float64) },
	}),
	"/": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}:     func(v1, v2 interface{}) interface{} { return v1.(int64) / v2.(int64) },
		two{reflect.Float64, reflect.Float64}: func(v1, v2 interface{}) interface{} { return v1.(float64) / v2.(float64) },
	}),
	"%": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) % v2.(int64) },
	}),
	"&": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) & v2.(int64) },
	}),
	"|": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) | v2.(int64) },
	}),
	"^": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) ^ v2.(int64) },
	}),
	"<<": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) << uint64(v2.(int64)) },
	}),
	">>": operatorBinary(binaryAction{
		two{reflect.Int64, reflect.Int64}: func(v1, v2 interface{}) interface{} { return v1.(int64) >> uint64(v2.(int64)) },
	}),
}

func (calc *Calc) Exec(data map[string]interface{}) (result interface{}, err error) {
	do := &does{make([]interface{}, 0, 16), data}
	if err = do.exec(calc.operators); err != nil {
		return
	} else if len(do.stack) != 1 {
		err = errorTable[2]
	} else {
		result = do.stack[0]
	}
	return
}

func (calc *Calc) ExecToSlice(data map[string]interface{}) (result []interface{}, err error) {
	do := &does{make([]interface{}, 0, 16), data}
	if err = do.exec(calc.operators); err != nil {
		result = do.stack
	}
	return
}

type unaryAction map[reflect.Kind]func(interface{}) interface{}

func operatorUnary(action unaryAction) operator {
	return func(do *does) error {
		if last := len(do.stack) - 1; last < 0 {
			return errorTable[0]
		} else if action, exists := action[reflect.TypeOf(do.stack[last]).Kind()]; !exists {
			return errorTable[1]
		} else {
			do.stack[last] = action(do.stack[last])
			return nil
		}
	}
}

type two [2]reflect.Kind

type binaryAction map[two]func(interface{}, interface{}) interface{}

func operatorBinary(action binaryAction) operator {
	return func(do *does) error {
		if last := len(do.stack) - 1; last <= 0 {
			return errorTable[0]
		} else if action, exists := action[two{reflect.TypeOf(do.stack[last-1]).Kind(), reflect.TypeOf(do.stack[last]).Kind()}]; !exists {
			return errorTable[1]
		} else {
			do.stack[last-1] = action(do.stack[last-1], do.stack[last])
			do.stack = do.stack[:last]
			return nil
		}
	}
}

func operatorConstant(value interface{}) operator {
	return func(do *does) error {
		do.stack = append(do.stack, value)
		return nil
	}
}

func operatorSelect(expressions [][]operator) operator {
	return func(do *does) error {
		if last := len(do.stack) - 1; last < 0 {
			return errorTable[0]
		} else if code, ok := do.stack[last].(int64); !ok {
			return errorTable[3]
		} else if code < 0 || code > int64(last) {
			return errorTable[4]
		} else {
			do.stack = do.stack[:last]
			return do.exec(expressions[code])
		}
	}
}

func operatorDrop(do *does) error {
	if last := len(do.stack); last == 0 {
		return errorTable[0]
	} else {
		do.stack = do.stack[:last-1]
		return nil
	}
}

func operatorDup(do *does) error {
	if last := len(do.stack); last == 0 {
		return errorTable[0]
	} else {
		do.stack = append(do.stack, do.stack[last-1])
		return nil
	}
}

func operatorSwap(do *does) error {
	if last := len(do.stack) - 1; last < 1 {
		return errorTable[0]
	} else {
		temp := do.stack[last]
		do.stack[last] = do.stack[last-1]
		do.stack[last-1] = temp
		return nil
	}
}

func operatorOver(do *does) error {
	if last := len(do.stack); last < 2 {
		return errorTable[0]
	} else {
		do.stack = append(do.stack, do.stack[last-2])
		return nil
	}
}

func operatorArgument(do *does) error {
	if last := len(do.stack) - 1; last < 1 {
		return errorTable[0]
	} else if name, ok := do.stack[last].(string); !ok {
		return errors.New("stack value is not string")
	} else if raw, exists := do.args[name]; !exists {
		return errors.New("argument not exists")
	} else if value, err := getArgument(raw); err != nil {
		return err
	} else {
		do.stack[last] = value
		return nil
	}
}

func getArgument(value interface{}) (interface{}, error) {
	switch value := reflect.Indirect(reflect.ValueOf(value)); value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	case reflect.String:
		return value.String(), nil
	default:
		return nil, errors.New("argument type is not valid")
	}
}
