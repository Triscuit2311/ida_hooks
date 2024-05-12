package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// il2cpp:0000000180C079C0 ; float __stdcall DriftController__GetDriftAngle(DriftController_o *this, CarX_Car_o *car, float *dotProduct, const MethodInfo *method)
// il2cpp:000000018077F4B0; void __stdcall Quests_QuestController__AddEnginePart(int32_t id, int32_t count, const MethodInfo* method)
//        ADDR / OFFSET     ret    callcv     methodname//funcname                          args

var knownTypes = []string{"char", "bool", "int", "float", "double", "wchar_t", "signed char", "short int", "unsigned short int", "unsigned int", "long int", "unsigned long int", "long long int", "unsigned long long int", "int8_t", "uint8_t", "int16_t", "uint16_t", "int32_t", "uint32_t", "int64_t", "uint64_t", "int_fast8_t", "uint_fast8_t", "int_fast16_t", "uint_fast16_t", "int_fast32_t", "uint_fast32_t", "int_fast64_t", "uint_fast64_t", "int_least8_t", "uint_least8_t", "int_least16_t", "uint_least16_t", "int_least32_t", "uint_least32_t", "int_least64_t", "uint_least64_t", "intptr_t", "uintptr_t", "intmax_t", "uintmax_t", "size_t", "ptrdiff_t", "wchar_t", "char16_t", "char32_t", "DWORD", "WORD", "BYTE", "LONG", "BOOL", "CHAR", "SHORT", "LONG", "FLOAT", "DOUBLE"}

type parameter struct {
	Name, Type                  string
	IsPtr, IsConst, IsBasicType bool
}

func (p parameter) GetParam() string {
	ret := ""
	if p.IsConst {
		ret += "const "
	}
	ret += p.Type
	if p.IsPtr {
		ret += "*"
	}
	ret += " "
	ret += p.Name
	return ret
}

func MakeParamsList(raw string) []parameter {
	params := make([]parameter, 0)
	paramsStrs := strings.Split(raw, ",")
	for _, p := range paramsStrs {
		p = strings.TrimSpace(p)
		isConst := strings.Contains(p, "const")
		if isConst {
			p = strings.TrimSpace(strings.Replace(p, "const", "", 1))
		}
		isPtr := strings.Contains(p, "*")
		if isPtr {
			p = strings.TrimSpace(strings.Replace(p, "*", "", 1))
		}
		pType := "void"
		pName := "none"
		isBasicType := false
		for _, v := range knownTypes {
			// only parse complete type (i.e. int32_t not parsed as int)
			if strings.HasPrefix(p, v+" ") || strings.HasPrefix(p, v+"*") {
				pType = v
				p = strings.TrimSpace(strings.Replace(p, v, "", 1))
				isBasicType = true
			}
		}
		if !isBasicType {
			pName = strings.Split(p, " ")[1]
		} else {
			pName = p
		}
		params = append(params, parameter{
			Name:        pName,
			Type:        pType,
			IsPtr:       isPtr,
			IsConst:     isConst,
			IsBasicType: isBasicType,
		})
	}
	return params
}

func Generate(str string) {

	t := strings.Split(str, ":")
	binarySection := t[0]
	_ = binarySection
	str = strings.TrimSpace(t[1])

	t = strings.Split(str, ";")
	addr := strings.TrimSpace(t[0])
	str = strings.TrimSpace(t[1])

	t = strings.Split(str, "(")
	rawParams := strings.TrimSuffix(strings.TrimSpace(t[1]), ")")
	str = strings.TrimSpace(t[0])

	t = strings.Split(str, " ")
	retType := strings.TrimSpace(t[0])
	callingConv := strings.TrimSpace(t[1])
	functionName := strings.TrimSpace(t[2])

	addrVal, _ := strconv.ParseInt(addr, 16, 64)
	offsetVal := addrVal - 0x0000000180000000

	params := MakeParamsList(rawParams)

	prettyFunctionName := strings.Replace(strings.Replace(functionName, "__", "::", 5), "_", "::", 5)

	// Print Comment data
	{
		fmt.Printf("// Method: %v\n", prettyFunctionName)
		fmt.Printf("// Addr: 0x%016X | Offset: 0x%016X\n", addrVal, offsetVal)
	}
	//Print Typedef
	{
		fmt.Printf("typedef %v (%v* _type_%v)(", retType, callingConv, functionName)
		for i, param := range params {
			fmt.Printf(param.GetParam())
			if i < len(params)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Print(");\n")
	}
	//Print Original Function Ptr Def
	{
		fmt.Printf("_type_%v o_%v{nullptr};\n", functionName, functionName)
	}
	//Print Hook function
	{
		fmt.Printf("%v %v hooked_%v(", retType, callingConv, functionName)
		for i, param := range params {
			fmt.Printf(param.GetParam())
			if i < len(params)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Print("){\n")
		fmt.Printf("\tLOG(\"%v called\")\n\t", prettyFunctionName)

		if !strings.HasPrefix(retType, "void") {
			fmt.Print("return ")
		}

		fmt.Printf("o_%v(", functionName)
		for i, param := range params {
			fmt.Print(param.Name)
			if i < len(params)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Print(");\n}\n")
	}
	//Print hook data struct
	{
		fmt.Printf("hook_data hk_%v = {\n\t(void*)hooked_%v,\n\t(void**)&o_%v,\n\t\"%v\",\n\t0x%016X\n};\n",
			functionName, functionName, functionName, prettyFunctionName, offsetVal)
	}

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ida_hooks <required: function sig>")
		fmt.Println("Example: ida_hooks \"section::000000018077F4B0; void __stdcall f_name(ComplexType * this, int a, float* b)\"")
		return
	}

	Generate(os.Args[1])
	//Generate("il2cpp:000000018077F4B0; void __stdcall Quests_QuestController__AddEnginePart(int32_t id, int32_t count, const MethodInfo* method)")
	//Generate("il2cpp:0000000180C079C0 ; float __stdcall DriftController__GetDriftAngle(DriftController_o *this, CarX_Car_o *car, float *dotProduct, const MethodInfo *method)")

}
