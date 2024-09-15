package gen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

type Result struct {
	OutputFile string
	Content    []byte
}

func (r *Result) WriteOutput() error {
	return os.WriteFile(r.OutputFile, r.Content, 0o644)
}

type Options struct {
}

func Generate(ctx context.Context, inputFile string, outputFile string, opts *Options) (*Result, error) {
	if opts == nil {
		opts = &Options{}
	}
	l := newLoader()
	pf, err := l.loadInputFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("Load Input file '%s' Failed:\n%w", inputFile, errors.Join(l.errs...))
	}
	absOutput, err := filepath.Abs(filepath.Clean(outputFile))
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path of %s: %w", outputFile, err)
	}
	tdata, err := l.generate(pf, absOutput)
	if err != nil {
		return nil, fmt.Errorf("generate types failed: %w", err)
	}
	var buf bytes.Buffer
	if err = generateOutput(*tdata, &buf); err != nil {
		return nil, fmt.Errorf("write output failed: %w", err)
	}
	goSrc := buf.Bytes()
	os.WriteFile(absOutput, goSrc, 0o644)
	fmtSrc, err := format.Source(goSrc)
	if err != nil {
		return nil, fmt.Errorf("format source failed: %w", err)
	}
	return &Result{
		OutputFile: absOutput,
		Content:    fmtSrc,
	}, nil

}

const (
	constructorPrefix = "Instrument"
	typePrefix        = "instrumented"
	funcPrefix        = "Trace"
)

func (l *loader) generate(file *ParsedFile, outFile string) (*TemplateData, error) {
	destDir := filepath.Dir(outFile)
	destPaths, err := getFullPackagePath(destDir)
	if err != nil {
		return nil, fmt.Errorf("could not get full package path for %s: %w", outFile, err)
	}
	if err = os.MkdirAll(destDir, os.FileMode(0o755)); err != nil {
		return nil, fmt.Errorf("could not create directory %s: %w", destDir, err)
	}
	pkgName := filepath.Base(destDir)
	if file.PkgPath == destPaths.packagePath {
		pkgName = file.Package
	}
	exportFile := TemplateData{
		Package: pkgName,
	}
	it := newTypeImporter(destPaths.packagePath, l)
	cache := newAutoSetterFuncCache(it, l)
	for _, fn := range file.Functions {
		fun, err := l.createWrapperFunction(file, fn, it, cache)
		if err != nil {
			return nil, err
		}
		exportFile.Functions = append(exportFile.Functions, fun)
	}
	if _, err = it.useType("github.com/justenwalker/genstrument", "SpanStarter"); err != nil {
		return nil, err
	}
	for _, iface := range file.Interfaces {
		var wi TemplateTypeConfig
		wi.Name = iface.Name.Name
		if iface.Config.ExternalType != nil {
			wi.ExternalType = it.resolveExpr(iface.Config.ExternalType)
		}
		wi.ConstructorName = prefix(wi.Name, iface.Config.ConstructorPrefix, constructorPrefix)
		wi.TypeName = prefix(wi.Name, iface.Config.Prefix, typePrefix)
		for _, f := range iface.Functions {
			fun, err := l.createWrapperFunction(file, f, it, cache)
			if err != nil {
				return nil, err
			}
			wi.Functions = append(wi.Functions, fun)
		}
		specs := it.TypeParams(iface.TypeParams)
		wi.TypeParamSpec = typeParamsToSpec(iface.TypeParams, specs)
		wi.TypeParamNames = typeParamNames(iface.TypeParams)
		if iface.Config.ExternalType != nil {
			wi.QualifiedName = it.resolveExpr(iface.Config.ExternalType)
		} else {
			wi.QualifiedName = it.resolveExpr(iface.Name)
		}
		exportFile.Types = append(exportFile.Types, wi)
	}
	exportFile.Imports = it.Imports()
	if err = errors.Join(l.errs...); err != nil {
		return nil, fmt.Errorf("could not generate file '%s': %w", outFile, err)
	}
	return &exportFile, nil
}

func prefix(name string, prefix string, def string) string {
	if prefix == "" {
		return fmt.Sprintf("%s%s", def, name)
	}
	return fmt.Sprintf("%s%s", prefix, name)

}

func typeParamNames(params []TypeParam) string {
	if len(params) == 0 {
		return ""
	}
	names := make([]string, len(params))
	for i, p := range params {
		names[i] = p.Name
	}
	return "[" + strings.Join(names, ",") + "]"
}

func typeParamsToSpec(params []TypeParam, specs []string) string {
	if len(params) == 0 {
		return ""
	}
	if len(params) != len(specs) {
		return ""
	}
	typeSpecs := make([]string, len(params))
	for i, p := range params {
		typeSpecs[i] = fmt.Sprintf("%s %s", p.Name, specs[i])
	}
	return "[" + strings.Join(typeSpecs, ",") + "]"
}

type argNameDisambiguator struct {
	names map[string]int
}

func newArgNameDisambiguator(reserved ...string) *argNameDisambiguator {
	names := make(map[string]int)
	for _, r := range reserved {
		names[r] = 0
	}
	return &argNameDisambiguator{
		names: names,
	}
}

func (d *argNameDisambiguator) disambiguate(name string) string {
	if n, ok := d.names[name]; ok {
		d.names[name] = n + 1
		return fmt.Sprintf("%s%d", name, n)
	}
	d.names[name] = 1
	return name
}

func (l *loader) findType(expr ast.Expr) types.Type {
	if expr == nil {
		return nil
	}
	switch t := expr.(type) {
	case *ast.Ident:
		typeName := t.Name
		if pkg, ok := l.typeNameToPackage[typeName]; ok {
			if obj := pkg.Types.Scope().Lookup(typeName); obj != nil {
				return obj.Type()
			}
			return nil
		}
		if found := types.Universe.Lookup(typeName); found != nil {
			return found.Type()
		}
		return nil
	case *ast.SelectorExpr:
		pkgName := t.X.(*ast.Ident).Name
		typeName := t.Sel.Name
		if pkg, ok := l.importNameToPackage[pkgName]; ok {
			return pkg.Types.Scope().Lookup(typeName).Type()
		}
		if pkgPath, ok := l.pkgNameToPkgPath[pkgName]; ok {
			if pkg, ok := l.pkgPathToPackage[pkgPath]; ok {
				return pkg.Types.Scope().Lookup(typeName).Type()
			}
		}
	}
	return nil
}

func (l *loader) createWrapperFunction(file *ParsedFile, f Function, it *typeImporter, cache *autoSetterFuncCache) (fun TemplateFunctionConfig, err error) {
	fun.Name = f.Name.Name
	if et := f.Config.ExternalType; et != nil {
		fun.QualifiedName, err = it.useSelector(f.Config.ExternalType)
	} else {
		fun.QualifiedName, err = it.useType(file.PkgPath, fun.Name)
	}

	fun.WrapperName = prefix(fun.Name, f.Config.Prefix, funcPrefix)
	if err != nil {
		return TemplateFunctionConfig{}, err
	}
	ctxArg := -1
	errArg := -1
	fun.OperationName = f.Config.OperationName
	typeSpecs := it.TypeParams(f.TypeParams)
	fun.TypeParamSpec = typeParamsToSpec(f.TypeParams, typeSpecs)
	fun.TypeParamNames = typeParamNames(f.TypeParams)
	d := newArgNameDisambiguator("span", "tr")
	fun.TracerArg = "tr"
	for i, a := range f.Arguments {
		var arg TemplateFunctionArg
		arg.Name = a.Name
		arg.Type = it.resolveExpr(a.Type)
		fun.ArgHasAttributes = fun.ArgHasAttributes || l.extractFuncArgument(&f, &arg, a, it, cache)
		if l.typeIsContext(a.Type) && ctxArg == -1 {
			arg.Name = "ctx"
			ctxArg = i
		}
		if arg.Name == "" {
			arg.Name = fmt.Sprintf("arg%d", i)
		}
		arg.Name = d.disambiguate(arg.Name)
		if i == ctxArg {
			fun.ContextArg = arg.Name
		}

		fun.Arguments = append(fun.Arguments, arg)
	}
	for i, a := range f.Returns {
		var arg TemplateFunctionArg
		arg.Name = a.Name
		arg.Type = it.resolveExpr(a.Type)
		if setter, ok := f.Config.AttributeFunctions[a.Name]; ok {
			arg.AttrKey = setter.Key
			if typ := l.findType(setter.Func); typ != nil {
				arg.AttrFunc = it.resolveExpr(setter.Func)
				fun.ReturnHasAttributes = true
			} else {
				if typ = l.findType(a.Type); typ != nil {
					arg.AttrFunc = cache.autoSetterFunc(typ)
				}
				fun.ReturnHasAttributes = fun.ReturnHasAttributes || arg.AttrFunc != ""
			}
		}
		if l.typeIsError(a.Type) && errArg == -1 {
			arg.Name = "err"
			errArg = i
		}
		if arg.Name == "" {
			arg.Name = fmt.Sprintf("ret%d", i)
		}
		arg.Name = d.disambiguate(arg.Name)
		if errArg == i {
			fun.ErrorReturn = arg.Name
		}
		fun.Returns = append(fun.Returns, arg)
	}
	return fun, nil
}

func (l *loader) extractFuncArgument(f *Function, arg *TemplateFunctionArg, a Arg, it *typeImporter, cache *autoSetterFuncCache) bool {
	setter, ok := f.Config.AttributeFunctions[a.Name]
	if !ok {
		return false
	}
	var typeParamName string
	if id, ok := a.Type.(*ast.Ident); ok {
		for _, tp := range f.TypeParams {
			if tp.Name == id.Name {
				typeParamName = tp.Name
			}
		}
	}
	if setter.Func != nil {
		arg.AttrFunc = it.resolveExpr(setter.Func)
		if arg.AttrFunc == "" {
			l.recordError(a.Type.Pos(), fmt.Errorf("could not resolve attribute function %s", setter.Func))
			return false
		}
		return true
	}
	if setter.Key == "" {
		return false
	}
	if typeParamName != "" {
		l.recordError(a.Type.Pos(), fmt.Errorf("cannot find auto-setter function for generic type %s", typeParamName))
		return false
	}
	typ := l.findType(a.Type)
	if typ == nil {
		l.recordError(a.Type.Pos(), fmt.Errorf("cannot find type %s", a.Type))
		return false
	}
	arg.AttrFunc = cache.autoSetterFunc(typ)
	if arg.AttrFunc == "" {
		l.recordError(a.Type.Pos(), fmt.Errorf("cannot find auto-setter function for type %s", a.Type))
		return false
	}
	return true
}

type autoSetterFuncCache struct {
	autoFuncMap map[types.Type]string
}

func newAutoSetterFuncCache(it *typeImporter, l *loader) *autoSetterFuncCache {
	var cache autoSetterFuncCache
	cache.autoFuncMap = make(map[types.Type]string)
	var (
		setString = selector("genstrument", "SetStringAttribute")
		setInt    = selector("genstrument", "SetIntAttribute")
		setBool   = selector("genstrument", "SetBoolAttribute")
		setFloat  = selector("genstrument", "SetFloatAttribute")
		setError  = selector("genstrument", "SetErrorAttribute")
	)
	l.importPackage("github.com/justenwalker/genstrument")
	var autoFuncMap = map[types.BasicKind]ast.Expr{
		types.String:  setString,
		types.Int64:   setInt,
		types.Int32:   setInt,
		types.Int16:   setInt,
		types.Int8:    setInt,
		types.Int:     setInt,
		types.Bool:    setBool,
		types.Float64: setFloat,
		types.Float32: setFloat,
	}
	if typ := l.findType(ident("error")); typ != nil {
		cache.autoFuncMap[typ] = it.resolveExpr(setError)
	}
	for k, v := range autoFuncMap {
		if typ := l.findType(v); typ != nil {
			cache.autoFuncMap[types.Typ[k]] = it.resolveExpr(v)
		}
	}
	return &cache
}

func selector(x string, sel string) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   &ast.Ident{Name: x},
		Sel: &ast.Ident{Name: sel},
	}
}

func ident(name string) *ast.Ident {
	return &ast.Ident{Name: name}
}

func (c *autoSetterFuncCache) autoSetterFunc(t types.Type) string {
	for k, v := range c.autoFuncMap {
		if types.AssignableTo(t, k) {
			return v
		}
		if types.ConvertibleTo(t, k) {
			return v
		}
	}
	return ""
}
