package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"go/doc"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/imports"
	gofumpt "mvdan.cc/gofumpt/format"
)

var (
	inputFile   string
	outputFile  string
	packageName string
	prefix      string
	suffix      string
)

func init() {
	flag.StringVar(&inputFile, "i", "", "Remote url or local path of the protocol xml file")
	flag.StringVar(&outputFile, "o", "", "Path of the generated output go file")
	flag.StringVar(&packageName, "pkg", "", "Go package name")
	flag.StringVar(&prefix, "prefix", "", "Specifiy prefix to trim")
	flag.StringVar(&suffix, "suffix", "", "Specifiy suffix to trim")
}

type Protocol struct {
	XMLName    xml.Name    `xml:"protocol"`
	Name       string      `xml:"name,attr"`
	Copyright  string      `xml:"copyright"`
	Interfaces []Interface `xml:"interface"`
}

type Interface struct {
	XMLName     xml.Name    `xml:"interface"`
	Name        string      `xml:"name,attr"`
	Description Description `xml:"description"`
	Requests    []Request   `xml:"request"`
	Events      []Event     `xml:"event"`
	Enums       []Enum      `xml:"enum"`
	Version     int         `xml:"version,attr"`
}

type Request struct {
	XMLName     xml.Name    `xml:"request"`
	Name        string      `xml:"name,attr"`
	Type        string      `xml:"type,attr"`
	Description Description `xml:"description"`
	Args        []Arg       `xml:"arg"`
	Since       int         `xml:"since,attr"`
}

type Event struct {
	XMLName     xml.Name    `xml:"event"`
	Name        string      `xml:"name,attr"`
	Type        string      `xml:"type,attr"`
	Description Description `xml:"description"`
	Args        []Arg       `xml:"arg"`
	Since       int         `xml:"since,attr"`
}

type Enum struct {
	XMLName     xml.Name    `xml:"enum"`
	Name        string      `xml:"name,attr"`
	Description Description `xml:"description"`
	Entries     []Entry     `xml:"entry"`
	Since       int         `xml:"since,attr"`
	Bitfield    bool        `xml:"bitfield,attr"`
}

type Entry struct {
	XMLName     xml.Name    `xml:"entry"`
	Name        string      `xml:"name,attr"`
	Value       string      `xml:"value,attr"`
	Summary     string      `xml:"summary,attr"`
	Description Description `xml:"description"`
	Since       int         `xml:"since,attr"`
}

type Arg struct {
	XMLName     xml.Name    `xml:"arg"`
	Name        string      `xml:"name,attr"`
	Type        string      `xml:"type,attr"`
	Summary     string      `xml:"summary,attr"`
	Interface   string      `xml:"interface,attr"`
	Enum        string      `xml:"enum,attr"`
	Description Description `xml:"description"`
	AllowNull   bool        `xml:"allow-null,attr"`
}

type Description struct {
	XMLName xml.Name `xml:"description"`
	Text    string   `xml:",chardata"`
	Summary string   `xml:"summary,attr"`
}

var protocol Protocol

func main() {
	flag.Parse()

	if inputFile == "" || outputFile == "" {
		flag.Usage()
		return
	}

	src, err := getInputFile(inputFile)
	if err != nil {
		log.Fatalf("unable to get input file: %v", err)
	}

	if err1 := xml.NewDecoder(src).Decode(&protocol); err1 != nil {
		log.Fatalf("unable to decode protocol xml: %v", err1)
	}
	if err2 := src.Close(); err2 != nil {
		log.Printf("unable to close input file: %v", err2)
	}

	w := &bytes.Buffer{}

	// Header
	fmt.Fprintf(w, "// Generated by go-wayland-scanner\n")
	fmt.Fprintf(w, "// https://github.com/hempflower/go-wayland/cmd/go-wayland-scanner\n")
	fmt.Fprintf(w, "// XML file : %s\n", inputFile)
	fmt.Fprintf(w, "//\n")
	fmt.Fprintf(w, "// %s Protocol Copyright: \n", protocol.Name)
	fmt.Fprint(w, comment(protocol.Copyright))
	fmt.Fprintf(w, "\n\n")
	fmt.Fprintf(w, "package %s\n", packageName)

	// Imports
	if protocol.Name != "wayland" {
		fmt.Fprintf(w, "import \"github.com/hempflower/go-wayland/wayland/client\"\n")
		fmt.Fprintf(w, "import xdg_shell \"github.com/hempflower/go-wayland/wayland/stable/xdg-shell\"\n")
	}

	// Intefaces
	for _, v := range protocol.Interfaces {
		writeInterface(w, v)
	}

	dst, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("unable to create output file: %v", err)
	}

	if _, err := dst.Write(fmtFile(w.Bytes())); err != nil {
		log.Fatalf("unable to copy to output file: %v", err)
	}

	if err := dst.Close(); err != nil {
		log.Fatalf("unable to close to output file: %v", err)
	}
}

func getInputFile(file string) (io.ReadCloser, error) {
	if strings.HasPrefix(file, "http") {
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return os.Open(file)
}

func fmtFile(b []byte) []byte {
	// Run goimports
	b, err := imports.Process("", b, nil)
	if err != nil {
		log.Fatalf("cannot run goimports on file: %v", err)
	}

	langVersion := ""
	out, err := exec.Command("go", "list", "-m", "-f", "{{.GoVersion}}").Output()
	outSlice := bytes.Split(out, []byte("\n"))
	out = outSlice[0]
	out = bytes.TrimSpace(out)
	if err == nil && len(out) > 0 {
		langVersion = string(out)
	}

	// Run gofumpt
	b, err = gofumpt.Source(b, gofumpt.Options{LangVersion: langVersion, ExtraRules: true})
	if err != nil {
		log.Fatalf("cannot run gofumpt on file: %v", err)
	}

	return b
}

var typeToGoTypeMap map[string]string = map[string]string{
	"int":    "int32",
	"uint":   "uint32",
	"fixed":  "float64",
	"string": "string",
	"object": "Proxy",
	"array":  "[]byte",
	"fd":     "int",
}

func writeInterface(w io.Writer, v Interface) {
	ifaceName := toCamel(v.Name)
	ifaceNameLower := toLowerCamel(v.Name)

	// Interface name constant
	fmt.Fprintf(w, "// %sName : %s\n", ifaceName, doc.Synopsis(v.Description.Summary))
	fmt.Fprintf(w, "const %sName = \"%s\"\n", ifaceName, v.Name)

	// Interface struct
	fmt.Fprintf(w, "// %s : %s\n", ifaceName, doc.Synopsis(v.Description.Summary))
	fmt.Fprint(w, comment(v.Description.Text))
	fmt.Fprintf(w, "type %s struct {\n", ifaceName)
	if protocol.Name != "wayland" {
		fmt.Fprintf(w, "client.BaseProxy\n")
	} else {
		fmt.Fprintf(w, "BaseProxy\n")
	}
	for _, event := range v.Events {
		fmt.Fprintf(w, "%sHandler %s%sHandlerFunc\n", toLowerCamel(event.Name), ifaceName, toCamel(event.Name))
	}
	fmt.Fprintf(w, "}\n")

	// Constructor
	fmt.Fprintf(w, "// New%s : %s\n", ifaceName, doc.Synopsis(v.Description.Summary))
	fmt.Fprint(w, comment(v.Description.Text))
	if protocol.Name != "wayland" {
		fmt.Fprintf(w, "func New%s(ctx *client.Context) *%s {\n", ifaceName, ifaceName)
	} else {
		fmt.Fprintf(w, "func New%s(ctx *Context) *%s {\n", ifaceName, ifaceName)
	}
	fmt.Fprintf(w, "%s := &%s{}\n", ifaceNameLower, ifaceName)
	fmt.Fprintf(w, "ctx.Register(%s)\n", ifaceNameLower)
	fmt.Fprintf(w, "return %s\n", ifaceNameLower)
	fmt.Fprintf(w, "}\n")

	// Requests
	for i, r := range v.Requests {
		writeRequest(w, ifaceName, i, r)
	}

	if !hasDestructor(v) {
		fmt.Fprintf(w, "func (i *%s) Destroy() (error) {\n", ifaceName)
		fmt.Fprintf(w, "i.Context().Unregister(i)\n")
		fmt.Fprintf(w, "return nil\n")
		fmt.Fprintf(w, "}\n")
	}

	// Enums
	for _, e := range v.Enums {
		writeEnum(w, ifaceName, e)
	}

	// Events
	for _, e := range v.Events {
		writeEvent(w, ifaceName, e)
	}

	// Event dispatcher
	writeEventDispatcher(w, ifaceName, v)
}

func writeRequest(w io.Writer, ifaceName string, opcode int, r Request) {
	requestName := toCamel(r.Name)

	// Generate param & returns types
	params := []string{}
	returnTypes := []string{}
	for _, arg := range r.Args {
		argNameLower := toLowerCamel(arg.Name)
		argIface := toCamel(arg.Interface)

		if !isLocalInterface(arg.Interface) {
			if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
				argIface = "client." + toCamelPrefix(arg.Interface, "wl_")
			} else if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
				argIface = "xdg_shell." + toCamelPrefix(arg.Interface, "xdg_")
			}
		}

		switch arg.Type {
		case "new_id":
			if arg.Interface != "" {
				returnTypes = append(returnTypes, "*"+argIface)
			} else {
				// Special for wl_registry.bind
				params = append(params, "iface string", "version uint32", "id Proxy")
			}

		case "object":
			params = append(params, argNameLower+" *"+argIface)

		case "int", "uint", "fixed",
			"string", "array", "fd":
			params = append(params, argNameLower+" "+typeToGoTypeMap[arg.Type])
		}
	}

	returnTypes = append(returnTypes, "error")

	fmt.Fprintf(w, "// %s : %s\n", requestName, doc.Synopsis(r.Description.Summary))
	fmt.Fprint(w, comment(r.Description.Text))
	fmt.Fprintf(w, "//\n")
	for _, arg := range r.Args {
		argNameLower := toLowerCamel(arg.Name)

		if arg.Summary != "" && arg.Type != "new_id" {
			fmt.Fprintf(w, "//  %s: %s\n", argNameLower, doc.Synopsis(arg.Summary))
		}
	}
	fmt.Fprintf(w, "func (i *%s) %s(%s) (%s) {\n", ifaceName, requestName, strings.Join(params, ","), strings.Join(returnTypes, ","))
	if r.Type == "destructor" {
		fmt.Fprintf(w, "defer i.Context().Unregister(i)\n")
	}

	// Create new objects, if any
	newObjects := []string{}
	for _, arg := range r.Args {
		if arg.Type == "new_id" && arg.Interface != "" {
			argNameLower := toLowerCamel(arg.Name)
			argIface := toCamel(arg.Interface)

			if isLocalInterface(arg.Interface) {
				fmt.Fprintf(w, "%s := New%s(i.Context())\n", argNameLower, argIface)
			} else {
				if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
					fmt.Fprintf(w, "%s := client.New%s(i.Context())\n", argNameLower, toCamelPrefix(arg.Interface, "wl_"))
				} else if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
					fmt.Fprintf(w, "%s := xdg_shell.New%s(i.Context())\n", argNameLower, toCamelPrefix(arg.Interface, "xdg_"))
				}
			}

			newObjects = append(newObjects, argNameLower)
		}
	}

	// Create request
	fmt.Fprintf(w, "const opcode = %d\n", opcode)

	// Calculate size
	sizes := []string{"8"}
	canBeConst := true
	for _, arg := range r.Args {
		argNameLower := toLowerCamel(arg.Name)

		switch arg.Type {
		case "new_id":
			if arg.Interface != "" {
				sizes = append(sizes, "4")
			} else {
				canBeConst = false
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "ifaceLen := PaddedLen(len(iface)+1)\n")
				} else {
					fmt.Fprintf(w, "ifaceLen := client.PaddedLen(len(iface)+1)\n")
				}
				sizes = append(sizes, "(4 + ifaceLen)", "4", "4")
			}

		case "object", "int", "uint", "fixed":
			sizes = append(sizes, "4")

		case "string":
			canBeConst = false
			if protocol.Name == "wayland" {
				fmt.Fprintf(w, "%sLen := PaddedLen(len(%s)+1)\n", argNameLower, argNameLower)
			} else {
				fmt.Fprintf(w, "%sLen := client.PaddedLen(len(%s)+1)\n", argNameLower, argNameLower)
			}
			sizes = append(sizes, fmt.Sprintf("(4 + %sLen)", argNameLower))

		case "array":
			canBeConst = false
			fmt.Fprintf(w, "%sLen := len(%s)\n", argNameLower, argNameLower)
			sizes = append(sizes, fmt.Sprintf("%sLen", argNameLower))
		}
	}

	if canBeConst {
		fmt.Fprintf(w, "const _reqBufLen =  %s\n", strings.Join(sizes, "+"))
		fmt.Fprintf(w, "var _reqBuf [_reqBufLen]byte\n")
	} else {
		fmt.Fprintf(w, "_reqBufLen := %s\n", strings.Join(sizes, "+"))
		fmt.Fprintf(w, "_reqBuf := make([]byte, _reqBufLen)\n")
	}

	fmt.Fprintf(w, "l := 0\n")
	if protocol.Name == "wayland" {
		fmt.Fprintf(w, "PutUint32(_reqBuf[l:4], i.ID())\n")
	} else {
		fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:4], i.ID())\n")
	}
	fmt.Fprintf(w, "l += 4\n")
	if protocol.Name == "wayland" {
		fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))\n")
	} else {
		fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4], uint32(_reqBufLen<<16|opcode&0x0000ffff))\n")
	}
	fmt.Fprintf(w, "l += 4\n")

	fdIndex := -1
	for i, arg := range r.Args {
		argNameLower := toLowerCamel(arg.Name)

		switch arg.Type {
		case "object":
			if arg.AllowNull {
				fmt.Fprintf(w, "if %s == nil {\n", argNameLower)
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], 0)\n")
				} else {
					fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4], 0)\n")
				}
				fmt.Fprintf(w, "l += 4\n")
				fmt.Fprintf(w, "} else {\n")
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], %s.ID())\n", argNameLower)
				} else {
					fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4], %s.ID())\n", argNameLower)
				}
				fmt.Fprintf(w, "l += 4\n")
				fmt.Fprintf(w, "}\n")
			} else {
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], %s.ID())\n", argNameLower)
				} else {
					fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4], %s.ID())\n", argNameLower)
				}
				fmt.Fprintf(w, "l += 4\n")
			}

		case "new_id":
			if arg.Interface != "" {
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], %s.ID())\n", argNameLower)
				} else {
					fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4], %s.ID())\n", argNameLower)
				}
				fmt.Fprintf(w, "l += 4\n")
			} else {
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "PutString(_reqBuf[l:l+(4 + ifaceLen)], iface, ifaceLen)\n")
				} else {
					fmt.Fprintf(w, "client.PutString(_reqBuf[l:l+(4 + ifaceLen)], iface, ifaceLen)\n")
				}
				fmt.Fprintf(w, "l += (4 + ifaceLen)\n")

				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], uint32(version))\n")
				} else {
					fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4],  uint32(version))\n")
				}
				fmt.Fprintf(w, "l += 4\n")

				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], id.ID())\n")
				} else {
					fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4], id.ID())\n")
				}
				fmt.Fprintf(w, "l += 4\n")
			}

		case "int", "uint":
			if protocol.Name == "wayland" {
				fmt.Fprintf(w, "PutUint32(_reqBuf[l:l+4], uint32(%s))\n", argNameLower)
			} else {
				fmt.Fprintf(w, "client.PutUint32(_reqBuf[l:l+4], uint32(%s))\n", argNameLower)
			}
			fmt.Fprintf(w, "l += 4\n")

		case "fixed":
			if protocol.Name == "wayland" {
				fmt.Fprintf(w, "PutFixed(_reqBuf[l:l+4], %s)\n", argNameLower)
			} else {
				fmt.Fprintf(w, "client.PutFixed(_reqBuf[l:l+4], %s)\n", argNameLower)
			}
			fmt.Fprintf(w, "l += 4\n")

		case "string":
			if protocol.Name == "wayland" {
				fmt.Fprintf(w, "PutString(_reqBuf[l:l+(4 + %sLen)], %s, %sLen)\n", argNameLower, argNameLower, argNameLower)
			} else {
				fmt.Fprintf(w, "client.PutString(_reqBuf[l:l+(4 + %sLen)], %s, %sLen)\n", argNameLower, argNameLower, argNameLower)
			}
			fmt.Fprintf(w, "l += (4 + %sLen)\n", argNameLower)

		case "array":
			if protocol.Name == "wayland" {
				fmt.Fprintf(w, "PutArray(_reqBuf[l:l+(4 + %sLen)], %s)\n", argNameLower, argNameLower)
			} else {
				fmt.Fprintf(w, "client.PutArray(_reqBuf[l:l+(4 + %sLen)], %s)\n", argNameLower, argNameLower)
			}
			fmt.Fprintf(w, "l += %sLen\n", argNameLower)

		case "fd":
			fdIndex = i
		}
	}

	if fdIndex != -1 {
		arg := r.Args[fdIndex]
		argNameLower := toLowerCamel(arg.Name)

		fmt.Fprintf(w, "oob := unix.UnixRights(int(%s))\n", argNameLower)

		if canBeConst {
			fmt.Fprintf(w, "err := i.Context().WriteMsg(_reqBuf[:], oob)\n")
		} else {
			fmt.Fprintf(w, "err := i.Context().WriteMsg(_reqBuf, oob)\n")
		}
	} else {
		if canBeConst {
			fmt.Fprintf(w, "err := i.Context().WriteMsg(_reqBuf[:], nil)\n")
		} else {
			fmt.Fprintf(w, "err := i.Context().WriteMsg(_reqBuf, nil)\n")
		}
	}

	fmt.Fprintf(w, "return %s\n", strings.Join(append(newObjects, "err"), ","))
	fmt.Fprintf(w, "}\n")
}

func writeEnum(w io.Writer, ifaceName string, e Enum) {
	enumName := toCamel(e.Name)

	fmt.Fprintf(w, "type %s%s uint32\n", ifaceName, enumName)

	fmt.Fprintf(w, "// %s%s : %s\n", ifaceName, enumName, doc.Synopsis(e.Description.Summary))
	fmt.Fprint(w, comment(e.Description.Text))
	fmt.Fprintf(w, "const (\n")
	for _, entry := range e.Entries {
		entryName := toCamel(entry.Name)

		if entry.Summary != "" {
			fmt.Fprintf(w, "// %s%s%s : %s\n", ifaceName, enumName, entryName, doc.Synopsis(entry.Summary))
		}
		fmt.Fprintf(w, "%s%s%s %s%s = %s\n", ifaceName, enumName, entryName, ifaceName, enumName, entry.Value)
	}
	fmt.Fprintf(w, ")\n")

	fmt.Fprintf(w, "func (e %s%s) Name() string {\n", ifaceName, enumName)
	fmt.Fprintf(w, "switch e {\n")
	for _, entry := range e.Entries {
		entryName := toCamel(entry.Name)

		fmt.Fprintf(w, "case %s%s%s:\n", ifaceName, enumName, entryName)
		fmt.Fprintf(w, "return \"%s\"\n", entry.Name)
	}
	fmt.Fprintf(w, "default:\n")
	fmt.Fprintf(w, "return \"\"\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "}\n")

	fmt.Fprintf(w, "func (e %s%s) Value() string {\n", ifaceName, enumName)
	fmt.Fprintf(w, "switch e {\n")
	for _, entry := range e.Entries {
		entryName := toCamel(entry.Name)

		fmt.Fprintf(w, "case %s%s%s:\n", ifaceName, enumName, entryName)
		fmt.Fprintf(w, "return \"%s\"\n", entry.Value)
	}
	fmt.Fprintf(w, "default:\n")
	fmt.Fprintf(w, "return \"\"\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "}\n")

	fmt.Fprintf(w, "func (e %s%s) String() string {\n", ifaceName, enumName)
	fmt.Fprintf(w, "return  e.Name() + \"=\" +  e.Value()\n")
	fmt.Fprintf(w, "}\n")
}

func writeEvent(w io.Writer, ifaceName string, e Event) {
	eventName := toCamel(e.Name)
	eventNameLower := toLowerCamel(e.Name)

	// Event struct
	fmt.Fprintf(w, "// %s%sEvent : %s\n", ifaceName, eventName, doc.Synopsis(e.Description.Summary))
	fmt.Fprint(w, comment(e.Description.Text))
	fmt.Fprintf(w, "type %s%sEvent struct {\n", ifaceName, eventName)
	for _, arg := range e.Args {
		argName := toCamel(arg.Name)

		if arg.Description.Summary != "" {
			fmt.Fprintf(w, "// %s %s\n", argName, doc.Synopsis(arg.Description.Summary))
		}
		fmt.Fprint(w, comment(arg.Description.Text))
		switch arg.Type {
		case "object", "new_id":
			if arg.Interface != "" {
				argIface := toCamel(arg.Interface)

				if !isLocalInterface(arg.Interface) {
					if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
						argIface = "client." + toCamelPrefix(arg.Interface, "wl_")
					} else if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
						argIface = "xdg_shell." + toCamelPrefix(arg.Interface, "xdg_")
					}
				}

				fmt.Fprintf(w, "%s *%s\n", argName, argIface)
			} else {
				fmt.Fprintf(w, "%s Proxy\n", argName)
			}

		case "int", "uint", "fixed",
			"string", "array", "fd":
			fmt.Fprintf(w, "%s %s\n", argName, typeToGoTypeMap[arg.Type])
		}
	}
	fmt.Fprintf(w, "}\n")

	// Event handler interface
	fmt.Fprintf(w, "type %s%sHandlerFunc func(%s%sEvent)\n", ifaceName, eventName, ifaceName, eventName)

	// Set handler
	fmt.Fprintf(w, "// Set%sHandler : sets handler for %s%sEvent\n", eventName, ifaceName, eventName)
	fmt.Fprintf(w, "func (i *%s) Set%sHandler(f %s%sHandlerFunc) {\n", ifaceName, eventName, ifaceName, eventName)
	fmt.Fprintf(w, "i.%sHandler = f\n", eventNameLower)
	fmt.Fprintf(w, "}\n")
}

func writeEventDispatcher(w io.Writer, ifaceName string, v Interface) {
	if len(v.Events) == 0 {
		return
	}

	fmt.Fprintf(w, "func (i *%s) Dispatch(opcode uint32, fd int, data []byte) {\n", ifaceName)
	fmt.Fprintf(w, "switch opcode {\n")
	for i, e := range v.Events {
		eventName := toCamel(e.Name)
		eventNameLower := toLowerCamel(e.Name)

		hasFd := false
		for _, arg := range e.Args {
			if arg.Type == "fd" {
				hasFd = true
				break
			}
		}

		fmt.Fprintf(w, "case %d:\n", i)
		fmt.Fprintf(w, "if i.%sHandler == nil {\n", eventNameLower)
		if hasFd {
			fmt.Fprintf(w, "if fd != -1 {\n")
			fmt.Fprintf(w, "unix.Close(fd)\n")
			fmt.Fprintf(w, "}\n")
		}
		fmt.Fprintf(w, "return\n")
		fmt.Fprintf(w, "}\n")

		fmt.Fprintf(w, "var e  %s%sEvent\n", ifaceName, eventName)

		if len(e.Args) > 0 && (len(e.Args) != 1 || e.Args[0].Type != "fd") {
			fmt.Fprintf(w, "l := 0\n")
		}

		for _, arg := range e.Args {
			argName := toCamel(arg.Name)
			argNameLower := toLowerCamel(arg.Name)

			switch arg.Type {
			case "object", "new_id":
				if arg.Interface != "" {
					argIface := toCamel(arg.Interface)

					if !isLocalInterface(arg.Interface) {
						if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
							argIface = "client." + toCamelPrefix(arg.Interface, "wl_")
						} else if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
							argIface = "xdg_shell." + toCamelPrefix(arg.Interface, "xdg_")
						}
					}

					fmt.Fprintf(w, "%s := &%s{}\n", argName, argIface)

					if protocol.Name == "wayland" {
						fmt.Fprintf(w, "i.Context().SetProxy(Uint32(data[l :l+4]), %s)\n", argName)
					} else {
						fmt.Fprintf(w, "i.Context().SetProxy(client.Uint32(data[l :l+4]), %s)\n", argName)
					}

					fmt.Fprintf(w, "e.%s = %s\n", argName, argName)
				} else {
					if protocol.Name == "wayland" {
						fmt.Fprintf(w, "e.%s = i.Context().GetProxy(Uint32(data[l :l+4]))\n", argName)
					} else {
						fmt.Fprintf(w, "e.%s = i.Context().GetProxy(client.Uint32(data[l :l+4]))\n", argName)
					}
				}
				fmt.Fprintf(w, "l += 4\n")

			case "fd":
				fmt.Fprintf(w, "e.%s = fd\n", argName)

			case "uint":
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "e.%s = Uint32(data[l : l+4])\n", argName)
				} else {
					fmt.Fprintf(w, "e.%s = client.Uint32(data[l : l+4])\n", argName)
				}
				fmt.Fprintf(w, "l += 4\n")

			case "int":
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "e.%s = int32(Uint32(data[l : l+4]))\n", argName)
				} else {
					fmt.Fprintf(w, "e.%s = int32(client.Uint32(data[l : l+4]))\n", argName)
				}
				fmt.Fprintf(w, "l += 4\n")

			case "fixed":
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "e.%s = Fixed(data[l : l+4])\n", argName)
				} else {
					fmt.Fprintf(w, "e.%s = client.Fixed(data[l : l+4])\n", argName)
				}
				fmt.Fprintf(w, "l += 4\n")

			case "string":
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "%sLen := PaddedLen(int(Uint32(data[l : l+4])))\n", argNameLower)
				} else {
					fmt.Fprintf(w, "%sLen := client.PaddedLen(int(client.Uint32(data[l : l+4])))\n", argNameLower)
				}
				fmt.Fprintf(w, "l += 4\n")
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "e.%s = String(data[l : l+%sLen])\n", argName, argNameLower)
				} else {
					fmt.Fprintf(w, "e.%s = client.String(data[l : l+%sLen])\n", argName, argNameLower)
				}
				fmt.Fprintf(w, "l += %sLen\n", argNameLower)

			case "array":
				if protocol.Name == "wayland" {
					fmt.Fprintf(w, "%sLen := int(Uint32(data[l : l+4]))\n", argNameLower)
				} else {
					fmt.Fprintf(w, "%sLen := int(client.Uint32(data[l : l+4]))\n", argNameLower)
				}
				fmt.Fprintf(w, "l += 4\n")
				fmt.Fprintf(w, "e.%s = make([]byte, %sLen)\n", argName, argNameLower)
				fmt.Fprintf(w, "copy(e.%s, data[l:l+%sLen])\n", argName, argNameLower)
				fmt.Fprintf(w, "l += %sLen\n", argNameLower)
			}
		}

		fmt.Fprintf(w, "\ni.%sHandler(e)\n", eventNameLower)
	}
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "}\n")
}

func toCamel(s string) string {
	s = strings.TrimPrefix(s, prefix)
	s = strings.TrimSuffix(s, suffix)
	s = strcase.ToCamel(s)
	return s
}

// Same as toCamel but with custom prefix to trim
func toCamelPrefix(s string, prefix string) string {
	s = strings.TrimPrefix(s, prefix)
	s = strcase.ToCamel(s)
	return s
}

func toLowerCamel(s string) string {
	s = strcase.ToLowerCamel(s)
	if s == "map" {
		s = "_map"
	}
	return s
}

func comment(s string) string {
	sb := strings.Builder{}

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		sb.WriteString("// ")
		sb.WriteString(strings.TrimSpace(scanner.Text()))
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "// \n")
}

func hasDestructor(v Interface) bool {
	for _, r := range v.Requests {
		if r.Type == "destructor" {
			return true
		}
	}

	return false
}

func isLocalInterface(iface string) bool {
	for _, v := range protocol.Interfaces {
		if v.Name == iface {
			return true
		}
	}

	return false
}
