# gotypeconverter [![](https://github.com/fuji8/gotypeconverter/workflows/build/badge.svg)](https://github.com/fuji8/gotypeconverter/actions)
gotypeconverter generates a function that converts two structurally different types.

[日本語（原本）](https://github.com/fuji8/gotypeconverter/blob/main/README_JA.md)


## Install
```shell
go install github.com/fuji8/gotypeconverter/cmd/gotypeconverter
```

## Usage
```shell
> gotypeconverter                          
gotypeconverter: gotypeconverter generates a function that converts two different named types.

Usage: gotypeconverter [-flag] [package]


Flags:
  -d string
        destination type
  -o string
        output file; if nil, output stdout
  -pkg string
        output package; if nil, the directoryName and packageName must be same and will be used
  -s string
        source type
  -structTag string
         (default "cvt")
```
## Caution
Make sure the directory name and package name are the same. If they cannot be the same, make the module name and package name the same and specify with -pkg.


## Examples
see [testdata](https://github.com/fuji8/gotypeconverter/tree/main/testdata/src)
### Basic example
```go basic.go
package main

type basicSrc struct {
	foo int
	bar float64
	x   string
	y   string
}

type basicDst struct {
	foo int
	bar float32
	x   string
	z   string `cvt:"y"`
}
```

```shell
> gotypeconverter -s basicSrc -d basicDst -pkg main  .
// Code generated by gotypeconverter; DO NOT EDIT.
package main

func ConvertbasicSrcTobasicDst(src basicSrc) (dst basicDst) {
        dst.foo = src.foo
        dst.x = src.x
        dst.z = src.y
        return
}
```

### Normal example
```go normal.go
package normal

type e struct {
	e       bool
	m       string
	x       int
	members []uint8
}

type ug struct {
	uaiueo uint8
	gaaaa  uint8
}

type normalSrc struct {
	x struct {
		A int
		B bool
		C string
		D string
	}
	y int
	z float64
	m string

	members []ug
}

type normalDst struct {
	e
	x string
	y int
	z float64
}
```

```bash
> gotypeconverter -s normalSrc -d normalDst .                                
// Code generated by gotypeconverter; DO NOT EDIT.
package normal

func ConvertnormalSrcTonormalDst(src normalSrc) (dst normalDst) {
        dst.e.m = src.m
        dst.e.x = src.x.A
        dst.e.members = make([]uint8, len(src.members))
        for i := range src.members {
                dst.e.members[i] = src.members[i].uaiueo
        }
        dst.x = src.x.C
        dst.y = src.y
        dst.z = src.z
        return
}
```

## Conversion rules
WIP

https://github.com/fuji8/gotypeconverter/blob/v0.1.0/gotypeconverter.go#L449-L543
```go
	switch dstT := dst.(type) {
	case *types.Basic:
		switch srcT := src.(type) {
		case *types.Basic:
		case *types.Named:
			return fm.otherAndNamed(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.otherAndSlice(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.otherAndStruct(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.otherAndPointer(dst, srcT, dstSelector, srcSelector, index, history)
		default:
		}


	case *types.Named:
		switch srcT := src.(type) {
		case *types.Basic:
			return fm.namedAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Named:
			return fm.namedAndNamed(dstT, srcT, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.namedAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.namedAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.namedAndOther(dstT, src, dstSelector, srcSelector, index, history)
		default:
			return fm.namedAndOther(dstT, src, dstSelector, srcSelector, index, history)
		}


	case *types.Slice:
		switch srcT := src.(type) {
		case *types.Basic:
			return fm.sliceAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Named:
			return fm.otherAndNamed(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.sliceAndSlice(dstT, srcT, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.sliceAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.otherAndPointer(dst, srcT, dstSelector, srcSelector, index, history)
		default:
			return fm.sliceAndOther(dstT, src, dstSelector, srcSelector, index, history)
		}


	case *types.Struct:
		switch srcT := src.(type) {
		case *types.Basic:
			return fm.structAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Named:
			return fm.otherAndNamed(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.otherAndSlice(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.structAndStruct(dstT, srcT, dstSelector, srcSelector, index, history) ||
				fm.structAndOther(dstT, src, dstSelector, srcSelector, index, history) ||
				fm.otherAndStruct(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.otherAndPointer(dst, srcT, dstSelector, srcSelector, index, history)
		default:
			return fm.structAndOther(dstT, src, dstSelector, srcSelector, index, history)
		}


	case *types.Pointer:
		switch srcT := src.(type) {
		case *types.Basic:
		case *types.Named:
			return fm.pointerAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.pointerAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.pointerAndOther(dstT, src, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.otherAndPointer(dst, srcT, dstSelector, srcSelector, index, history)
		default:
			return fm.pointerAndOther(dstT, src, dstSelector, srcSelector, index, history)
		}


	default:
		switch srcT := src.(type) {
		case *types.Basic:
		case *types.Named:
			return fm.otherAndNamed(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Slice:
			return fm.otherAndSlice(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Struct:
			return fm.otherAndStruct(dst, srcT, dstSelector, srcSelector, index, history)
		case *types.Pointer:
			return fm.otherAndPointer(dst, srcT, dstSelector, srcSelector, index, history)
		default:
		}


	}
```