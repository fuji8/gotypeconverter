package gotypeconverter

import (
	"flag"
	"fmt"
	"go/types"
	"io"
	"os"
	"regexp"
	"strings"

	ana "github.com/fuji8/gotypeconverter/analysis"
	"github.com/fuji8/gotypeconverter/ui"
	"golang.org/x/tools/go/packages"
)

const (
	doc = "gotypeconverter generates a function that converts two different named types."
)

var (
	flagOutput  string
	FlagVersion bool

	flagSrc, flagDst, FlagStructTag string
)

func init() {
	Gen.Flags.StringVar(&flagOutput, "o", "", "output file; if nil, output stdout")
	Gen.Flags.StringVar(&flagSrc, "s", "", "source type")
	Gen.Flags.StringVar(&flagDst, "d", "", "destination type")
	Gen.Flags.BoolVar(&FlagVersion, "v", false, "version")
	Gen.Flags.StringVar(&FlagStructTag, "structTag", "cvt", "")
}

// A Generator describes a code generator function and its options.
type Generator struct {
	// The Name of the generator must be a valid Go identifier
	// as it may appear in command-line flags, URLs, and so on.
	Name string

	// Doc is the documentation for the generator.
	// The part before the first "\n\n" is the title
	// (no capital or period, max ~60 letters).
	Doc string

	// Flags defines any flags accepted by the generator.
	// The manner in which these flags are exposed to the user
	// depends on the driver which runs the generator.
	Flags flag.FlagSet

	// Run applies the generator to a package.
	// It returns an error if the generator failed.
	//
	// To pass analysis results of depended analyzers between packages (and thus
	// potentially between address spaces), use Facts, which are
	// serializable.
	Run func([]*packages.Package) (string, error)

	Output func(pkg *types.Package) io.Writer
}

var Gen = &Generator{
	Name: "gotypeconverter",
	Doc:  doc,
	Run:  run,
}

func TypeOf4(pkg *types.Package, pkgname, typename string) types.Type {
	if typename == "" {
		return nil
	}

	if typename[0] == '*' {
		obj := TypeOf4(pkg, pkgname, typename[1:])
		if obj == nil {
			return nil
		}
		return types.NewPointer(obj)
	}

	if typename[0] == '[' {
		obj := TypeOf4(pkg, pkgname, typename[2:])
		if obj == nil {
			return nil
		}
		return types.NewSlice(obj)
	}
	if pkgname == "" {
		obj := pkg.Scope().Lookup(typename)
		return obj.Type()
	}

	for i := 0; i < pkg.Scope().NumChildren(); i++ {
		pkgN, ok := pkg.Scope().Child(i).Lookup(pkgname).(*types.PkgName)
		if ok {
			obj := pkgN.Imported().Scope().Lookup(typename)
			if obj == nil {
				return nil
			}
			return obj.Type()
		}
	}
	return nil
}

func splitToPkgAndType(s string) (string, string) {
	if idx := strings.LastIndex(s, "."); idx > 0 {
		pkgname := s[:idx]
		typename := s[idx+1:]
		jdxs := regexp.MustCompile(`^[\[\]\*]+`).FindStringIndex(pkgname)
		if jdxs != nil {
			typename = pkgname[:jdxs[1]] + typename
			pkgname = pkgname[jdxs[1]:]
		}
		return pkgname, typename
	}
	return "", s
}

func run(pkgs []*packages.Package) (string, error) {
	var (
		srcType, dstType types.Type
	)
	pkgIdx := 0
	for i, pkg := range pkgs {
		srcPkgName, srcTypeName := splitToPkgAndType(flagSrc)
		dstPkgName, dstTypeName := splitToPkgAndType(flagDst)
		srcType = TypeOf4(pkg.Types, srcPkgName, srcTypeName)
		dstType = TypeOf4(pkg.Types, dstPkgName, dstTypeName)
		if srcType != nil && dstType != nil {
			pkgIdx = i
			break
		}
	}

	if srcType == nil || dstType == nil {
		return "", fmt.Errorf("srcType or dstType is nil, srcType: %s, dstType: %s", srcType, dstType)
	}

	funcMaker := ana.InitFuncMaker(pkgs[pkgIdx].Types)
	funcMaker.MakeFunc(ana.InitType(dstType, flagDst), ana.InitType(srcType, flagSrc))

	if flagOutput == "" {
		src, err := ui.NoInfoGeneration(funcMaker)
		if err != nil {
			return "", err
		}
		return src, nil
	}

	src, err := ui.FileNameGeneration(funcMaker, flagOutput)
	if err != nil {
		return "", err
	}

	f, err := os.Create(flagOutput)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fmt.Fprint(f, src)

	return "", nil
}
