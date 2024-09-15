package gen

import (
	"go/ast"
)

type ParsedFile struct {
	Package    string
	PkgPath    string
	Interfaces []Interface
	Functions  []Function
}

type FunctionType string

type FunctionConfig struct {
	OperationName      string
	Prefix             string
	ExternalType       *ast.SelectorExpr
	AttributeFunctions map[string]*AttributeKeyFunc
}

type AttributeKeyFunc struct {
	Key  string
	Func ast.Expr
}

type Function struct {
	Name       *ast.Ident
	TypeParams []TypeParam
	Arguments  []Arg
	Returns    []Arg
	Config     FunctionConfig
}

type InterfaceConfig struct {
	Prefix            string
	ExternalType      *ast.SelectorExpr
	ConstructorPrefix string
}

type Interface struct {
	Name       *ast.Ident
	TypeParams []TypeParam
	Config     InterfaceConfig
	Functions  []Function
}

type TypeParam struct {
	Name     string
	TypeExpr ast.Expr
}

type Arg struct {
	Name string
	Type ast.Expr
}

type TemplateImport struct {
	Name    string
	Package string
	PkgPath string
}

type TemplateData struct {
	Package   string
	Imports   []TemplateImport
	Functions []TemplateFunctionConfig
	Types     []TemplateTypeConfig
}

type TemplateFunctionConfig struct {
	Name                string
	WrapperName         string
	QualifiedName       string
	OperationName       string
	TypeParamSpec       string
	TypeParamNames      string
	TracerArg           string
	ContextArg          string
	ErrorReturn         string
	ArgHasAttributes    bool
	ReturnHasAttributes bool
	Arguments           []TemplateFunctionArg
	Returns             []TemplateFunctionArg
}

type TemplateFunctionArg struct {
	Name     string
	Type     string
	AttrFunc string
	AttrKey  string
}

type TemplateTypeConfig struct {
	Name            string
	ExternalType    string
	TypeName        string
	ConstructorName string
	QualifiedName   string
	TypeParamSpec   string
	TypeParamNames  string
	Config          InterfaceConfig
	Functions       []TemplateFunctionConfig
}
