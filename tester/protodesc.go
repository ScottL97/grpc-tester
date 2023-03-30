package tester

import (
	"errors"
	"fmt"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"path/filepath"
	"strings"
)

func getMethodDescFromProto(method, proto string, importPaths []string) *desc.MethodDescriptor {
	p := &protoparse.Parser{ImportPaths: importPaths}
	if filepath.IsAbs(proto) {
		proto = filepath.Base(proto)
	}
	fmt.Println(proto)
	fds, err := p.ParseFiles(proto)
	if err != nil {
		panic(err)
	}

	fileDesc := fds[0]
	files := map[string]*desc.FileDescriptor{}
	files[fileDesc.GetName()] = fileDesc
	return getMethodDesc(method, files)
}

func getMethodDesc(call string, files map[string]*desc.FileDescriptor) *desc.MethodDescriptor {
	svc, mth, err := parseServiceMethod(call)
	if err != nil {
		panic(err)
	}

	dsc, err := findServiceSymbol(files, svc)
	if err != nil {
		panic(err)
	}
	if dsc == nil {
		panic(fmt.Errorf("cannot find service %q", svc))
	}

	sd, ok := dsc.(*desc.ServiceDescriptor)
	if !ok {
		panic(fmt.Errorf("cannot find service %q", svc))
	}

	mtd := sd.FindMethodByName(mth)
	if mtd == nil {
		panic(fmt.Errorf("service %q does not include a method named %q", svc, mth))
	}

	return mtd
}

func findServiceSymbol(resolved map[string]*desc.FileDescriptor, fullyQualifiedName string) (desc.Descriptor, error) {
	for _, fd := range resolved {
		if dsc := fd.FindSymbol(fullyQualifiedName); dsc != nil {
			return dsc, nil
		}
	}
	return nil, fmt.Errorf("cannot find service %q", fullyQualifiedName)
}

var errNoMethodNameSpecified = errors.New("no method name specified")

// parseServiceMethod parses the fully-qualified service name without a leading "."
// and the method name from the input string.
//
// valid inputs:
//
//	package.Service.Method
//	.package.Service.Method
//	package.Service/Method
//	.package.Service/Method
func parseServiceMethod(svcAndMethod string) (string, string, error) {
	if len(svcAndMethod) == 0 {
		return "", "", errNoMethodNameSpecified
	}
	if svcAndMethod[0] == '.' {
		svcAndMethod = svcAndMethod[1:]
	}
	if len(svcAndMethod) == 0 {
		return "", "", errNoMethodNameSpecified
	}
	switch strings.Count(svcAndMethod, "/") {
	case 0:
		pos := strings.LastIndex(svcAndMethod, ".")
		if pos < 0 {
			return "", "", newInvalidMethodNameError(svcAndMethod)
		}
		return svcAndMethod[:pos], svcAndMethod[pos+1:], nil
	case 1:
		split := strings.Split(svcAndMethod, "/")
		return split[0], split[1], nil
	default:
		return "", "", newInvalidMethodNameError(svcAndMethod)
	}
}

func newInvalidMethodNameError(svcAndMethod string) error {
	return fmt.Errorf("method name must be package.Service.Method or package.Service/Method: %q", svcAndMethod)
}
