package gen

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log"
	"sort"
	"strings"
)

type typeImporter struct {
	currentPackage string
	loader         *loader
	importsUsed    map[string]struct{}
	importsNamed   map[string]int
	packageToName  map[string]string
}

func newTypeImporter(pkgPath string, loader *loader) *typeImporter {
	return &typeImporter{
		currentPackage: pkgPath,
		loader:         loader,
		importsUsed:    make(map[string]struct{}),
		importsNamed:   make(map[string]int),
		packageToName:  make(map[string]string),
	}
}

func (it *typeImporter) useType(packagePath string, typeName string) (string, error) {
	pkg, err := it.loader.importPackage(packagePath)
	if err != nil {
		return "", err
	}
	pkgName := it.namePackage(pkg)
	return fmt.Sprintf("%s.%s", pkgName, typeName), nil
}

func (it *typeImporter) toQualifiedName(expr ast.Expr) (string, error) {
	var pkg *packages.Package
	var typeName string
	switch t := expr.(type) {
	case *ast.Ident:
		typeName = t.Name
		var ok bool
		pkg, ok = it.loader.typeNameToPackage[typeName]
		if !ok { // built-in type
			return typeName, nil
		}
	case *ast.SelectorExpr:
		id := t.X.(*ast.Ident)
		pkgName := id.Name
		typeName = t.Sel.Name
		var ok bool
		pkg, ok = it.loader.importNameToPackage[pkgName]
		if !ok {
			if pkgPath, ok := it.loader.pkgNameToPkgPath[pkgName]; ok {
				return it.useType(pkgPath, typeName)
			}
			//return pkgName, fmt.Errorf("package %s is not impoted", pkgName)
		}
	default:
		return "", fmt.Errorf("unexpected qualified name type %T", expr)
	}
	pkgName := it.namePackage(pkg)
	return fmt.Sprintf("%s.%s", pkgName, typeName), nil
}

func (it *typeImporter) namePackage(pkg *packages.Package) string {
	if name, ok := it.packageToName[pkg.PkgPath]; ok { // already named
		return name
	}
	it.importsUsed[pkg.PkgPath] = struct{}{}
	pkgName := pkg.Name
	switch pkgName {
	case "err", "ctx", "span": // don't use common var names as package names
		pkgName = "_" + pkgName
	}
	name := pkgName
	for {
		if i, ok := it.importsNamed[name]; ok { // someone has our desired name
			it.importsNamed[name] = i + 1
			name = fmt.Sprintf("%s%d", pkgName, i+1) // increment and try again
			continue
		}
		// found a valid name
		it.importsNamed[name] = 0
		it.packageToName[pkg.PkgPath] = name
		return name
	}
}

func (it *typeImporter) Imports() []TemplateImport {
	imports := make([]TemplateImport, 0, len(it.packageToName))
	for pkg, name := range it.packageToName {
		imports = append(imports, TemplateImport{
			Name:    name,
			Package: it.loader.pkgPathToPackage[pkg].Name,
			PkgPath: pkg,
		})
	}
	sort.Slice(imports, func(i, j int) bool {
		a, b := imports[i], imports[j]
		return strings.Compare(a.PkgPath, b.PkgPath) < 0
	})
	return imports
}

func (it *typeImporter) TypeParams(tps []TypeParam) []string {
	strs := make([]string, len(tps))
	for i, tp := range tps {
		strs[i] = it.resolveExpr(tp.TypeExpr)
	}
	return strs
}

func (it *typeImporter) resolveExpr(e ast.Expr) string {
	var sb strings.Builder
	it.resolveTypeSpec(&sb, e)
	re := sb.String()
	return re
}

func (it *typeImporter) resolveTypeSpec(sb *strings.Builder, expr ast.Expr) {
	if expr == nil {
		return
	}
	//log.Printf("resolveTypeSpec: %v <%[1]T>", expr)
	switch t := expr.(type) {
	case *ast.BasicLit:
		sb.WriteString(t.Value)
	case *ast.SelectorExpr: // foo.Package
		qual, _ := it.toQualifiedName(expr)
		sb.WriteString(qual)
		return
	case *ast.Ident: // any
		qual, _ := it.toQualifiedName(expr)
		sb.WriteString(qual)
	case *ast.StarExpr: // *T
		sb.WriteString("*")
		it.resolveTypeSpec(sb, t.X)
	case *ast.UnaryExpr: // ~string
		sb.WriteString(t.Op.String())
		it.resolveTypeSpec(sb, t.X)
	case *ast.ArrayType:
		sb.WriteByte('[')
		if t.Len != nil {
			it.resolveTypeSpec(sb, t.Len)
		}
		sb.WriteByte(']')
		it.resolveTypeSpec(sb, t.Elt)
	case *ast.MapType:
		sb.WriteString("map[")
		it.resolveTypeSpec(sb, t.Key)
		sb.WriteByte(']')
		it.resolveTypeSpec(sb, t.Value)
	case *ast.IndexExpr:
		it.resolveTypeSpec(sb, t.X)
		sb.WriteByte('[')
		it.resolveTypeSpec(sb, t.Index)
		sb.WriteByte(']')
	case *ast.BinaryExpr:
		it.resolveTypeSpec(sb, t.X)
		sb.WriteString(t.Op.String())
		it.resolveTypeSpec(sb, t.Y)
	case *ast.InterfaceType:
		sb.WriteString("interface{")
		if t.Methods != nil {
			for _, m := range t.Methods.List {
				if m.Type != nil {
					it.resolveTypeSpec(sb, m.Type)
				}
			}
		}
		sb.WriteByte('}')
	default:
		log.Printf("Unexpected Expression Type: %v %[1]T", t)
		return
	}
}
