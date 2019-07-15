package scalc

import (
	"fmt"
	"testing"
)

type rounds struct {
	expr    string
	res     []interface{}
	isError bool
}

func TestParseEmpty(t *testing.T) {
	for _, err := range test([]rounds{
		{"", []interface{}{""}, false},
		{" \n \r \v \t ", []interface{}{""}, false},
	}, nil) {
		t.Error(err)
	}
}

func TestParseConstant(t *testing.T) {
	for _, err := range test([]rounds{
		{" \\sa\\n\\tb\\c\\\\d\\se\\ '-1.5e+13 ' -3 0 75 1.25 -2.5 3e7 4e-8 0.0",
			[]interface{}{
				" a\n\tbc\\d e\\", "-1.5e+13", "",
				int64(-3), int64(0), int64(75),
				float64(1.25), float64(-2.5), float64(3e7), float64(4e-8), float64(0.0),
			}, false},
	}, nil) {
		t.Error(err)
	}
}

func TestStackOperations(t *testing.T) {
	for _, err := range test([]rounds{
		{"drop", nil, true},
		{"1 drop", []interface{}{}, false},
		{"1 2 drop", []interface{}{int64(1)}, false},
		{"dup", nil, true},
		{"1 dup", []interface{}{int64(1), int64(1)}, false},
		{"1 2 dup", []interface{}{int64(1), int64(2), int64(2)}, false},
		{"swap", nil, true},
		{"1 swap", nil, true},
		{"1 2 swap", []interface{}{int64(2), int64(1)}, false},
		{"1 2 3 swap", []interface{}{int64(1), int64(3), int64(2)}, false},
		{"over", nil, true},
		{"1 over", nil, true},
		{"1 2 over", []interface{}{int64(1), int64(2), int64(1)}, false},
		{"1 2 3 over", []interface{}{int64(1), int64(2), int64(3), int64(2)}, false},
	}, nil) {
		t.Error(err)
	}
}

func TestBinaryOperationErrors(t *testing.T) {
	for _, err := range test([]rounds{
		{"1 drop +", nil, true},
		{"1 +", nil, true},
	}, nil) {
		t.Error(err)
	}
}

func TestUnaryOperationErrors(t *testing.T) {
	for _, err := range test([]rounds{
		{"1 drop --", nil, true},
	}, nil) {
		t.Error(err)
	}
}

func TestSwitchSyntaxError(t *testing.T) {
	for _, test := range []string{
		"[",
		"]",
		";",
		"0 [ 1 2 + [ 3 4 - ]",
		"0 [ 1 2 + [ 3 4 - ] ] 2 5 ] 7 8",
		"0 ; [ 1 2 + [ 3 4 - ]",
		"0 [ 1 2 + [ 3 4 - ] 11 ; 12",
	} {
		if _, err := New(test); err == nil {
			t.Errorf("string %#v is parsed", test)
		}
	}
}

func TestSwitch(t *testing.T) {
	for _, err := range test([]rounds{
		{"[ ]", nil, true},
		{" -1 [ ]", nil, true},
		{" 1 [ ]", nil, true},
		{" -1 [ ; ]", nil, true},
		{" 2 [ ; ]", nil, true},
		{" 0 [ 1 0 / ]", nil, true},
		{" 0 [ 0 [ 1 0 / ] ]", nil, true},
		{"0 [ ]", []interface{}{}, false},
		{"0 [ ; ]", []interface{}{}, false},
		{"1 [ ; ]", []interface{}{}, false},
		{"0 [ 7 ; ]", []interface{}{int64(7)}, false},
		{"1 [ 7 ; ]", []interface{}{}, false},
		{"0 [ ; 7 ]", []interface{}{}, false},
		{"1 [ ; 7 ]", []interface{}{int64(7)}, false},
		{"0 [ 1 2 + ]", []interface{}{int64(3)}, false},
		{"0 [ 1 2 + ; 4 5 * ]", []interface{}{int64(3)}, false},
		{"1 [ 1 2 + ; 4 5 * ]", []interface{}{int64(20)}, false},
		{"0 [ 101 ; 102 ; 103 ] 1 [ 201 ; 202 ; 203 ] 2 [ 301 ; 302 ; 303 ]",
			[]interface{}{int64(101), int64(202), int64(303)}, false},
		{"0 [ 1 2 + [ 21 ; 22 ; 23 ; 24 ] ; 3 2 - [ 31 ; 32 ; 33 ; 34 ] ]",
			[]interface{}{int64(24)}, false},
		{"1 [ 1 2 + [ 21 ; 22 ; 23 ; 24 ] ; 3 2 - [ 31 ; 32 ; 33 ; 34 ] ]",
			[]interface{}{int64(32)}, false},
	}, nil) {
		t.Error(err)
	}
}

func TestCompareOperations(t *testing.T) {
	for _, err := range test([]rounds{
		{" 1 2 = 2 1 = 2 2 = ", []interface{}{int64(0), int64(0), int64(1)}, false},
		{" 1 2 != 2 1 != 2 2 != ", []interface{}{int64(1), int64(1), int64(0)}, false},
		{" 1 2 > 2 1 > 2 2 > ", []interface{}{int64(0), int64(1), int64(0)}, false},
		{" 1 2 < 2 1 < 2 2 < ", []interface{}{int64(1), int64(0), int64(0)}, false},
		{" 1 2 >= 2 1 >= 2 2 >= ", []interface{}{int64(0), int64(1), int64(1)}, false},
		{" 1 2 <= 2 1 <= 2 2 <= ", []interface{}{int64(1), int64(0), int64(1)}, false},
		{" 1.2 2.1 = 2.3 1.4 = 2.5 2.5 = ", []interface{}{int64(0), int64(0), int64(1)}, false},
		{" 1.2 2.1 != 2.3 1.4 != 2.5 2.5 != ", []interface{}{int64(1), int64(1), int64(0)}, false},
		{" 1.2 2.1 > 2.3 1.4 > 2.5 2.5 > ", []interface{}{int64(0), int64(1), int64(0)}, false},
		{" 1.2 2.1 < 2.3 1.4 < 2.5 2.5 < ", []interface{}{int64(1), int64(0), int64(0)}, false},
		{" 1.2 2.1 >= 2.3 1.4 >= 2.5 2.5 >= ", []interface{}{int64(0), int64(1), int64(1)}, false},
		{" 1.2 2.1 <= 2.3 1.4 <= 2.5 2.5 <= ", []interface{}{int64(1), int64(0), int64(1)}, false},
		{" aaa bbb = ddd ccc = eee eee = ", []interface{}{int64(0), int64(0), int64(1)}, false},
		{" aaa bbb != ddd ccc != eee eee != ", []interface{}{int64(1), int64(1), int64(0)}, false},
		{" aaa bbb > ddd ccc > eee eee > ", []interface{}{int64(0), int64(1), int64(0)}, false},
		{" aaa bbb < ddd ccc < eee eee < ", []interface{}{int64(1), int64(0), int64(0)}, false},
		{" aaa bbb >= ddd ccc >= eee eee >= ", []interface{}{int64(0), int64(1), int64(1)}, false},
		{" aaa bbb <= ddd ccc <= eee eee <= ", []interface{}{int64(1), int64(0), int64(1)}, false},
	}, nil) {
		t.Error(err)
	}
}

func TestUnaryOperators(t *testing.T) {
	for _, err := range test([]rounds{
		{"aaa int", nil, true},
		{"'1.25 int", nil, true},
		{"aaa float", nil, true},
		{"1.25e+12.7 float", nil, true},
		{"aaa string 125 string -7 string -1.25 string 1e3 string", []interface{}{"aaa", "125", "-7", "-1.25", "1000"}, false},
		{"'37 int '-9 int -7 int -1.99 int 4.99 int", []interface{}{int64(37), int64(-9), int64(-7), int64(-1), int64(4)}, false},
		{"37 float -9 float -7.75 float 12,35 float '-12.35 float '3e-7 float", []interface{}{float64(37.0), float64(-9.0), float64(-7.75), float64(12.35), float64(-12.35), float64(3e-7)}, false},
		{"25 -- 0 -- -25 --", []interface{}{int64(-25), int64(0), int64(25)}, false},
		{"25.25 -- 0.0 -- -25.25 --", []interface{}{float64(-25.25), float64(0.0), float64(25.25)}, false},
		{"25 ~ 0 ~ -25 ~", []interface{}{int64(^25), int64(^0), int64(^-25)}, false},
		{"25 ! 0 ! -25 !", []interface{}{int64(0), int64(1), int64(0)}, false},
		{"25 abs 0 abs -25 abs", []interface{}{int64(25), int64(0), int64(25)}, false},
		{"25.25 abs 0.0 abs -25.25 abs", []interface{}{float64(25.25), float64(0.0), float64(25.25)}, false},
		{"25 sign 0 sign -25 sign", []interface{}{int64(1), int64(0), int64(-1)}, false},
		{"25.25 sign 0.0 sign -25.25 sign", []interface{}{int64(1), int64(0), int64(-1)}, false},
		{"' len aaa len", []interface{}{int64(0), int64(3)}, false},
		{"2.25 sqrt", []interface{}{float64(1.5)}, false},
		{"1.0 0.0 / isInf -1.0 0.0 / isInf -1.0 isInf 0.0 isInf 1.0 isInf", []interface{}{int64(1), int64(1), int64(0), int64(0), int64(0)}, false},
		{"-1.0 sqrt isNaN -1.0 isNaN 0.0 isNaN 1.0 isNaN", []interface{}{int64(1), int64(0), int64(0), int64(0)}, false},
	}, nil) {
		t.Error(err)
	}
}

func TestBinaryOperators(t *testing.T) {
	for _, err := range test([]rounds{
		{"1 0 /", nil, true},
		{"'1 0 %", nil, true},
		{"1 -3 +", []interface{}{int64(-2)}, false},
		{"1.25 -3.5 +", []interface{}{float64(-2.25)}, false},
		{"aaa bbb +", []interface{}{"aaabbb"}, false},
		{"1 -3 -", []interface{}{int64(4)}, false},
		{"1.25 -3.5 -", []interface{}{float64(4.75)}, false},
		{"2 -3 *", []interface{}{int64(-6)}, false},
		{"1.25 -3.5 *", []interface{}{float64(-4.375)}, false},
		{"5 -3 /", []interface{}{int64(-1)}, false},
		{"16.0 0.5 /", []interface{}{float64(32.0)}, false},
		{"5 3 % 5 -3 % -5 -3 % -5 3 %", []interface{}{int64(2), int64(2), int64(-2), int64(-2)}, false},
		{"13 25 &", []interface{}{int64(9)}, false},
		{"13 25 |", []interface{}{int64(29)}, false},
		{"13 25 ^", []interface{}{int64(20)}, false},
		{"13 2 << -13 2 <<", []interface{}{int64(52), int64(-52)}, false},
		{"13 2 >> -13 2 >>", []interface{}{int64(3), int64(-4)}, false},
	}, nil) {
		t.Error(err)
	}
}

func TestAttributes(t *testing.T) {
	pi := int(106)
	pi8 := int8(107)
	pi16 := int16(108)
	pi32 := int32(109)
	pi64 := int64(110)
	pu64 := uint64(20)
	pf32 := float32(221.25)
	pf64 := float64(222.5)
	pst := "baaaa"

	for _, err := range test([]rounds{
		{"@", []interface{}{""}, true},
		{"test @", []interface{}{""}, true},
		{"u_64 @", nil, true},
		{"pu_64 @", nil, true},
		{"nu_64 @", nil, true},
		{"i_ @", []interface{}{int64(6)}, false},
		{"i_8 @", []interface{}{int64(7)}, false},
		{"i_16 @", []interface{}{int64(8)}, false},
		{"i_32 @", []interface{}{int64(9)}, false},
		{"i_64 @", []interface{}{int64(10)}, false},
		{"f_32 @", []interface{}{float64(21.5)}, false},
		{"f_64 @", []interface{}{float64(22.25)}, false},
		{"st @", []interface{}{"aaaa"}, false},
		{"pi_ @", []interface{}{int64(106)}, false},
		{"pi_8 @", []interface{}{int64(107)}, false},
		{"pi_16 @", []interface{}{int64(108)}, false},
		{"pi_32 @", []interface{}{int64(109)}, false},
		{"pi_64 @", []interface{}{int64(110)}, false},
		{"pf_32 @", []interface{}{float64(221.25)}, false},
		{"pf_64 @", []interface{}{float64(222.5)}, false},
		{"pst @", []interface{}{"baaaa"}, false},
	}, map[string]interface{}{
		"i_":    int(6),
		"i_8":   int8(7),
		"i_16":  int16(8),
		"i_32":  int32(9),
		"i_64":  int64(10),
		"u_64":  uint64(20),
		"f_32":  float32(21.5),
		"f_64":  float64(22.25),
		"st":    "aaaa",
		"pi_":   &pi,
		"pi_8":  &pi8,
		"pi_16": &pi16,
		"pi_32": &pi32,
		"pi_64": &pi64,
		"pu_64": &pu64,
		"pf_32": &pf32,
		"pf_64": &pf64,
		"pst":   &pst,
	}) {
		t.Error(err)
	}
}

func TestCalculators_Exec(t *testing.T) {
	for _, test := range []rounds{
		{"drop", nil, true},
		{"1 drop", nil, true},
		{"1 2", nil, true},
		{"", []interface{}{""}, false},
		{"7", []interface{}{int64(7)}, false},
		{"data @", []interface{}{int64(125)}, false},
	} {
		if calc, err := New(test.expr); err != nil {
			t.Errorf("string %#v parse => %#v", test.expr, err)
		} else if res, err := calc.Exec(map[string]interface{}{"data": 125}); (err != nil) != test.isError {
			if test.isError {
				t.Errorf("string %#v calculate %#v is not error", test.expr, res)
			} else {
				t.Errorf("string %#v calculate => %#v", test.expr, err)
			}
		} else if err == nil && res != test.res[0] {
			t.Errorf("string %#v calculate result %#v != %#v", test.expr, res, test.res)
		}
	}
}

func test(check []rounds, data map[string]interface{}) (result []string) {
	result = make([]string, 0)
	for _, test := range check {
		if calc, err := New(test.expr); err != nil {
			result = append(result, fmt.Sprintf("string %#v parse => %#v", test.expr, err))
		} else if res, err := calc.ExecToSlice(data); err != nil {
			if !test.isError {
				result = append(result, fmt.Sprintf("string %#v calculate => %#v", test.expr, err))
			}
		} else if test.isError {
			result = append(result, fmt.Sprintf("string %#v calculate %#v is not error", test.expr, res))
		} else if len(res) != len(test.res) {
			result = append(result, fmt.Sprintf("string %#v calculate result %#v != %#v", test.expr, res, test.res))
		} else {
			for key, val := range res {
				if val != test.res[key] {
					result = append(result, fmt.Sprintf("string %#v calculate result %#v != %#v", test.expr, res, test.res))
					break
				}
			}
		}
	}
	return
}
