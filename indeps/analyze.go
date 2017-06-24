package indeps

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/loader"
)

func analyze(pkg *loader.PackageInfo) (*Graph, error) {
	g := NewGraph()

	// First we will add all of our top-level decls to the graph as nodes
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			switch td := decl.(type) {
			case *ast.GenDecl:
				for _, declSpec := range td.Specs {
					switch ts := declSpec.(type) {
					case *ast.ValueSpec:
						for _, ident := range ts.Names {
							var node Node
							switch td.Tok {
							case token.CONST:
								node = Constant(ident.Name)
							default:
								node = Variable(ident.Name)
							}
							g.AddNode(node)
						}
					case *ast.TypeSpec:
						node := Type(ts.Name.Name)
						g.AddNode(node)
					case *ast.ImportSpec:
						// don't care
					default:
						return nil, fmt.Errorf("unknown ast.Spec type %T", ts)
					}
				}
			case *ast.FuncDecl:
				// We're only interested in global functions, since we'll
				// treat methods as belonging to their receiver type.
				if td.Recv == nil {
					node := Function(td.Name.Name)
					g.AddNode(node)
				}
			default:
				return nil, fmt.Errorf("unknown ast.Decl type %T", td)
			}
		}
	}

	// Now we'll walk again and peep into the bodies of the top-level decls
	// to see what they reference.
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			switch td := decl.(type) {
			case *ast.GenDecl:
				for _, declSpec := range td.Specs {
					switch ts := declSpec.(type) {
					case *ast.ValueSpec:
						for _, ident := range ts.Names {
							addDefEdges(ident, g, pkg)
						}
					case *ast.TypeSpec:
						addDefEdges(ts.Name, g, pkg)
					case *ast.ImportSpec:
					// don't care
					default:
						return nil, fmt.Errorf("unknown ast.Spec type %T", ts)
					}
				}
			case *ast.FuncDecl:
				addFuncEdges(td, g, pkg)
			default:
				return nil, fmt.Errorf("unknown ast.Decl type %T", td)
			}
		}
	}

	return g, nil
}

func addDefEdges(ident *ast.Ident, g *Graph, pkg *loader.PackageInfo) {
	obj := pkg.Defs[ident]

	switch to := obj.(type) {
	case *types.Const:
		fromNode := Constant(to.Name())
		fromPkg := to.Pkg()

		addTypeEdges(fromNode, fromPkg, to.Type(), g, pkg)

	case *types.Var:
		fromNode := Variable(to.Name())
		fromPkg := to.Pkg()

		if nt, isNamed := to.Type().(*types.Named); isNamed {
			name := nt.Obj()
			toPkg := name.Pkg()
			if fromPkg == toPkg {
				toNode := Type(name.Name())
				g.AddEdge(fromNode, toNode)
			}
		}

	case *types.TypeName:
		fromNode := Type(to.Name())
		fromPkg := to.Pkg()

		ty := to.Type().Underlying()

		addTypeEdges(fromNode, fromPkg, ty, g, pkg)

	case *types.Func:

	default:
		panic(fmt.Sprintf("unsupported definition object %T", obj))
	}
}

func addFuncEdges(f *ast.FuncDecl, g *Graph, pkg *loader.PackageInfo) {
	fo := pkg.Defs[f.Name].(*types.Func)

	sig := fo.Type().(*types.Signature)
	var fromNode Node
	fromPkg := fo.Pkg()
	if recv := sig.Recv(); recv != nil {
		// Method is represented by its receiver type
		recvType := recv.Type()
		if _, isNamed := recvType.(*types.Named); !isNamed {
			recvType = recvType.(*types.Pointer).Elem()
		}
		fromNode = Type(recvType.(*types.Named).Obj().Name())
	} else {
		// Global function represents itself
		fromNode = Function(fo.Name())
	}

	addTypeEdges(fromNode, fromPkg, sig, g, pkg)

	// Now we'll look at the function body. This is a trickier affair,
	// since there's all sorts of ways a function body can refer to
	// other symbols in this package.
	visitor := astVisitor(func(n ast.Node) {
		switch tn := n.(type) {
		case *ast.Ident:
			if tv, ok := pkg.Types[tn]; ok {
				addTypeEdges(fromNode, fromPkg, tv.Type, g, pkg)
			}
			if obj := pkg.Defs[tn]; obj != nil {
				addTypeEdges(fromNode, fromPkg, obj.Type(), g, pkg)
			}
			if obj := pkg.Uses[tn]; obj != nil {
				addTypeEdges(fromNode, fromPkg, obj.Type(), g, pkg)
			}
			if obj := pkg.Implicits[tn]; obj != nil {
				addTypeEdges(fromNode, fromPkg, obj.Type(), g, pkg)
			}
		}
	})
	ast.Walk(visitor, f.Body)
}

func addTypeEdges(fromNode Node, fromPkg *types.Package, ty types.Type, g *Graph, pkg *loader.PackageInfo) {

	switch tt := ty.(type) {
	case *types.Named:
		name := tt.Obj()
		toPkg := name.Pkg()
		if fromPkg == toPkg {
			toNode := Type(name.Name())
			g.AddEdge(fromNode, toNode)
		}
	case *types.Struct:
		for i := 0; i < tt.NumFields(); i++ {
			f := tt.Field(i)
			addTypeEdges(fromNode, fromPkg, f.Type(), g, pkg)
		}
	case *types.Interface:
		for i := 0; i < tt.NumEmbeddeds(); i++ {
			addTypeEdges(fromNode, fromPkg, tt.Embedded(i), g, pkg)
		}
		for i := 0; i < tt.NumMethods(); i++ {
			fn := tt.Method(i)
			addTypeEdges(fromNode, fromPkg, fn.Type(), g, pkg)
		}
	case *types.Signature:
		params := tt.Params()
		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			addTypeEdges(fromNode, fromPkg, param.Type(), g, pkg)
		}

		results := tt.Results()
		for i := 0; i < results.Len(); i++ {
			result := results.At(i)
			addTypeEdges(fromNode, fromPkg, result.Type(), g, pkg)
		}

		if tt.Recv() != nil {
			addTypeEdges(fromNode, fromPkg, tt.Recv().Type(), g, pkg)
		}

	case *types.Map:
		addTypeEdges(fromNode, fromPkg, tt.Key(), g, pkg)
		addTypeEdges(fromNode, fromPkg, tt.Elem(), g, pkg)
	case *types.Slice:
		addTypeEdges(fromNode, fromPkg, tt.Elem(), g, pkg)
	case *types.Pointer:
		addTypeEdges(fromNode, fromPkg, tt.Elem(), g, pkg)
	case *types.Array:
		addTypeEdges(fromNode, fromPkg, tt.Elem(), g, pkg)
	case *types.Chan:
		addTypeEdges(fromNode, fromPkg, tt.Elem(), g, pkg)
	case *types.Basic:
		// basic types can't belong to our package, so ignore
	default:
		panic(fmt.Sprintf("unsupported type kind %T", ty))
	}

}

type astVisitor func(node ast.Node)

func (v astVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		v(node)
		return v
	}
	return nil
}
