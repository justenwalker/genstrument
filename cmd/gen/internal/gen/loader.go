package gen

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"strconv"
)

type loader struct {
	fset                *token.FileSet
	pkg                 *packages.Package
	pkgPathToPackage    map[string]*packages.Package
	pkgPathToImport     map[string]importInfo
	typeNameToPackage   map[string]*packages.Package
	pkgNameToPkgPath    map[string]string
	importNameToPackage map[string]*packages.Package
	errs                []error
}

type importInfo struct {
	alias string
	name  string
}

func newLoader() *loader {
	return &loader{
		pkgPathToPackage:    make(map[string]*packages.Package),
		pkgPathToImport:     make(map[string]importInfo),
		typeNameToPackage:   make(map[string]*packages.Package),
		pkgNameToPkgPath:    make(map[string]string),
		importNameToPackage: make(map[string]*packages.Package),
	}
}

func (l *loader) loadInputFile(inputFile string) (*ParsedFile, error) {
	absInput, err := filepath.Abs(filepath.Clean(inputFile))
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path of %s: %w", inputFile, err)
	}
	l.fset = token.NewFileSet()
	cfg := &packages.Config{
		Fset: l.fset,
		Mode: packages.NeedTypesInfo | packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedImports | packages.NeedDeps | packages.NeedCompiledGoFiles,
	}
	pkgs, err := packages.Load(cfg, "file="+absInput)
	if err != nil {
		return nil, fmt.Errorf("error loading packages: %w", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected exactly one package, got %d", len(pkgs))
	}
	return l.loadPackage(absInput, pkgs[0])
}

func (l *loader) importPackage(pkgPath string) (*packages.Package, error) {
	pkg, ok := l.pkgPathToPackage[pkgPath]
	if ok {
		return pkg, nil
	}
	cfg := &packages.Config{
		Fset: l.fset,
		Mode: packages.NeedTypesInfo | packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedImports | packages.NeedDeps | packages.NeedCompiledGoFiles,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		l.pkgNameToPkgPath[pkg.Name] = pkg.PkgPath
		l.pkgPathToPackage[pkg.PkgPath] = pkg
	}
	return l.pkgPathToPackage[pkgPath], nil
}

func (l *loader) loadPackage(filename string, pkg *packages.Package) (*ParsedFile, error) {
	l.pkg = pkg
	l.pkgPathToPackage[pkg.PkgPath] = pkg
	l.loadTypes(pkg)
	var file *ast.File
	for i, f := range pkg.CompiledGoFiles {
		if f == filename {
			file = pkg.Syntax[i]
			break
		}
	}
	if file == nil {
		return nil, fmt.Errorf("could not get compiled go file for '%s'", filename)
	}
	for _, imp := range pkg.Imports {
		l.pkgPathToPackage[imp.PkgPath] = imp
		l.pkgNameToPkgPath[imp.Name] = imp.PkgPath
		l.pkgPathToImport[imp.PkgPath] = importInfo{
			name:  imp.Name,
			alias: imp.Name,
		}
	}
	for _, imp := range file.Imports {
		var name string
		if imp.Name != nil {
			name = imp.Name.Name
		}
		pkgPath, _ := strconv.Unquote(imp.Path.Value)
		info := l.pkgPathToImport[pkgPath]
		switch name {
		case ".": // dot import
			info.alias = "."
			l.loadTypes(pkg.Imports[pkgPath])
		case "_": // anonymous imports
			info.alias = "_"
		case "": // use existing package name
			l.importNameToPackage[info.name] = pkg.Imports[pkgPath]
		default: // renamed
			info.alias = name
			l.importNameToPackage[name] = pkg.Imports[pkgPath]
		}
		l.pkgPathToImport[pkgPath] = info
	}
	parsedFile := ParsedFile{
		Package: pkg.Name,
		PkgPath: pkg.PkgPath,
	}

	for _, d := range file.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			if decl.Tok != token.TYPE {
				continue // only care about type declarations
			}
			if decl.Doc == nil {
				continue // undocumented types don't count
			}
			l.loadGenDecl(&parsedFile, decl)
		case *ast.FuncDecl:
			if decl.Doc == nil {
				continue // undocumented functions don't count
			}
			l.loadFuncDecl(&parsedFile, decl)
		}
	}
	if err := errors.Join(l.errs...); err != nil {
		return nil, fmt.Errorf("could not load file '%s': %w", filename, err)
	}
	return &parsedFile, nil
}

func (l *loader) loadTypes(pkg *packages.Package) {
	scope := pkg.Types.Scope()
	names := scope.Names()
	for _, typeName := range names {
		l.typeNameToPackage[typeName] = pkg
	}
}

func (l *loader) loadGenDecl(file *ParsedFile, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		l.loadTypeSpec(file, typeSpec, decl.Doc)
	}
}

func (l *loader) loadTypeSpec(file *ParsedFile, spec *ast.TypeSpec, doc *ast.CommentGroup) {
	cfg, ok := l.toInterfaceConfig(doc)
	if !ok {
		return // not documented with interface marker
	}
	typeDef, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		l.recordError(typeDef.Pos(), fmt.Errorf("type is not an interface"))
		return
	}
	iface, err := l.loadInterface(spec, typeDef, cfg)
	if err != nil {
		return
	}
	file.Interfaces = append(file.Interfaces, iface)
	return
}

func (l *loader) loadInterface(spec *ast.TypeSpec, typeDef *ast.InterfaceType, cfg InterfaceConfig) (Interface, error) {
	var iface Interface
	if typeDef.Methods == nil || len(typeDef.Methods.List) == 0 {
		return iface, fmt.Errorf("interface has no methods")
	}
	iface.Name = spec.Name
	iface.Config = cfg
	if spec.TypeParams != nil {
		iface.TypeParams = l.extractTypeParams(spec.TypeParams)
	}
	for _, method := range typeDef.Methods.List {
		funcType := method.Type.(*ast.FuncType)
		if err := l.loadInterfaceMethod(&iface, method, funcType); err != nil {
			return iface, err
		}
	}
	return iface, nil
}

func (l *loader) extractTypeParams(fieldList *ast.FieldList) []TypeParam {
	if fieldList == nil {
		return nil
	}
	params := make([]TypeParam, 0, len(fieldList.List))
	for _, param := range fieldList.List {
		tp := TypeParam{
			Name:     param.Names[0].Name,
			TypeExpr: param.Type,
		}
		params = append(params, tp)
	}
	return params
}

func (l *loader) loadInterfaceMethod(iface *Interface, method *ast.Field, funcType *ast.FuncType) error {
	name := method.Names[0]
	fcfg, _ := l.toFunctionConfig(method.Doc)
	fn := l.loadFunction(iface, name, funcType, fcfg)
	if fn.Name == nil {
		return fmt.Errorf("failed to load function '%s'", name)
	}
	iface.Functions = append(iface.Functions, fn)
	return nil
}

func (l *loader) loadFunction(iface *Interface, name *ast.Ident, ft *ast.FuncType, cfg FunctionConfig) (fun Function) {
	fun.Name = name
	fun.Config = cfg
	if fun.Config.OperationName == "" {
		fun.Config.OperationName = fmt.Sprintf("%s:%s", l.pkg.Name, name)
		if iface != nil {
			fun.Config.OperationName = fmt.Sprintf("%s.%s:%s", l.pkg.Name, iface.Name, name)
		}
	}
	var typeParams []TypeParam
	if ft.TypeParams != nil {
		typeParams = l.extractTypeParams(ft.TypeParams)
		fun.TypeParams = typeParams
	}
	if iface != nil {
		typeParams = iface.TypeParams
	}
	if ft.Params == nil || len(ft.Params.List) == 0 {
		l.recordError(ft.Pos(), fmt.Errorf("function has no arguments: must have at least context. context in the first position"))
		return Function{}
	}
	if ft.Params != nil {
		for _, p := range ft.Params.List {
			arg, err := l.loadArgument(p, fun.TypeParams)
			if err != nil {
				l.recordError(p.Pos(), fmt.Errorf("bad function argument '%s': %w", p.Names, err))
				return Function{}
			}
			fun.Arguments = append(fun.Arguments, arg)
		}
	}

	if ft.Results != nil {
		for _, p := range ft.Results.List {
			arg, err := l.loadArgument(p, fun.TypeParams)
			if err != nil {
				l.recordError(p.Pos(), fmt.Errorf("bad function return value '%s': %w", p.Names, err))
				return Function{}
			}
			fun.Returns = append(fun.Returns, arg)
		}
	}
	return fun
}

func (l *loader) recordError(pos token.Pos, err error) {
	position := l.fset.Position(pos)
	err = fmt.Errorf("%s:%d: %w", position.Filename, position.Line, err)
	l.errs = append(l.errs, err)
}

func (l *loader) loadArgument(f *ast.Field, tps []TypeParam) (Arg, error) {
	var arg Arg
	if len(f.Names) != 0 && f.Names[0] != nil {
		arg.Name = f.Names[0].Name
	}
	arg.Type = f.Type
	return arg, nil
}

func (l *loader) loadFuncDecl(file *ParsedFile, decl *ast.FuncDecl) {
	fcfg, ok := l.toFunctionConfig(decl.Doc)
	if !ok {
		return
	}
	fun := l.loadFunction(nil, decl.Name, decl.Type, fcfg)
	if fun.Name != nil {
		file.Functions = append(file.Functions, fun)
	}
}

func (l *loader) typeIsContext(t ast.Expr) bool {
	sel, ok := t.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgName := sel.X.(*ast.Ident).Name
	if sel.Sel.Name != "Context" {
		return false
	}
	pkg, ok := l.pkgPathToPackage[pkgName]
	if !ok {
		return false
	}
	return pkg.PkgPath == "context"
}

func (l *loader) typeIsError(t ast.Expr) bool {
	id, ok := t.(*ast.Ident)
	if !ok {
		return false
	}
	return id.Name == "error"
}
