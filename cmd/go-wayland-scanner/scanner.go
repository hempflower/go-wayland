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

	protocol Protocol
)

func init() {
	flag.StringVar(&inputFile, "i", "", "Remote url or local path of the protocol xml file")
	flag.StringVar(&outputFile, "o", "", "Path of the generated output go file")
	flag.StringVar(&packageName, "pkg", "", "Go package name")
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
	fmt.Fprintf(w, "// https://github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner\n")
	fmt.Fprintf(w, "// XML file : %s\n", inputFile)
	fmt.Fprintf(w, "//\n")
	fmt.Fprintf(w, "// %s Protocol Copyright: \n", toCamel(protocol.Name))
	fmt.Fprint(w, comment(protocol.Copyright))
	fmt.Fprintf(w, "\n\n")
	fmt.Fprintf(w, "package %s\n", packageName)

	// Imports
	fmt.Fprintf(w, "import \"sync\"\n")
	if protocol.Name != "wayland" {
		fmt.Fprintf(w, "import \"github.com/rajveermalviya/go-wayland/client\"\n")
		fmt.Fprintf(w, "import xdg_shell \"github.com/rajveermalviya/go-wayland/stable/xdg-shell\"\n")
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

var (
	typeToGoTypeMap map[string]string = map[string]string{
		"int":    "int32",
		"uint":   "uint32",
		"fixed":  "float32",
		"string": "string",
		"object": "Proxy",
		"array":  "[]int32",
		"fd":     "uintptr",
	}
	typeToGetterMap map[string]string = map[string]string{
		"int":    "Int32",
		"uint":   "Uint32",
		"fixed":  "Float32",
		"string": "String",
		"array":  "Array",
		"fd":     "FD",
	}
)

func writeInterface(w io.Writer, v Interface) {
	ifaceName := toCamel(v.Name)
	ifaceNameLower := toLowerCamel(v.Name)

	// Interface struct
	fmt.Fprintf(w, "// %s : %s\n", ifaceName, doc.Synopsis(v.Description.Summary))
	fmt.Fprint(w, comment(v.Description.Text))
	fmt.Fprintf(w, "type %s struct {\n", ifaceName)
	if protocol.Name != "wayland" {
		fmt.Fprintf(w, "client.BaseProxy\n")
	} else {
		fmt.Fprintf(w, "BaseProxy\n")
	}
	if len(v.Events) > 0 {
		fmt.Fprintf(w, "mu sync.RWMutex\n")
	}
	for _, event := range v.Events {
		fmt.Fprintf(w, "%sHandlers []%s%sHandler\n", toLowerCamel(event.Name), ifaceName, toCamel(event.Name))
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

func writeRequest(w io.Writer, ifaceName string, order int, r Request) {
	requestName := toCamel(r.Name)

	params := []string{}
	reqParams := []string{}
	returnTypes := []string{}
	for _, arg := range r.Args {
		argNameLower := toLowerCamel(arg.Name)
		argIface := toCamel(arg.Interface)

		if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
			argIface = "client." + argIface
		}
		if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
			argIface = "xdg_shell." + argIface
		}

		switch arg.Type {
		case "new_id":
			if arg.Interface != "" {
				reqParams = append(reqParams, argNameLower)
				returnTypes = append(returnTypes, "*"+argIface)
			} else {
				// Special for wl_registry.bind
				params = append(params, "iface string", "version uint32", "id Proxy")
				reqParams = append(reqParams, "iface", "version", "id")
			}

		case "object":
			params = append(params, argNameLower+" *"+argIface)
			reqParams = append(reqParams, argNameLower)

		case "int", "uint", "fixed",
			"string", "array", "fd":
			params = append(params, argNameLower+" "+typeToGoTypeMap[arg.Type])
			reqParams = append(reqParams, argNameLower)
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
	newObjects := []string{}
	for _, arg := range r.Args {
		if arg.Type == "new_id" && arg.Interface != "" {
			argNameLower := toLowerCamel(arg.Name)
			argIface := toCamel(arg.Interface)

			if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
				fmt.Fprintf(w, "%s := client.New%s(i.Context())\n", argNameLower, argIface)
			} else if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
				fmt.Fprintf(w, "%s := xdg_shell.New%s(i.Context())\n", argNameLower, argIface)
			} else {
				fmt.Fprintf(w, "%s := New%s(i.Context())\n", argNameLower, argIface)
			}

			newObjects = append(newObjects, argNameLower)
		}
	}
	fmt.Fprintf(w, "err := i.Context().SendRequest(i, %d, %s)\n", order, strings.Join(reqParams, ","))
	fmt.Fprintf(w, "return %s\n", strings.Join(append(newObjects, "err"), ","))
	fmt.Fprintf(w, "}\n")
}

func writeEnum(w io.Writer, ifaceName string, e Enum) {
	enumName := toCamel(e.Name)

	fmt.Fprintf(w, "// %s%s : %s\n", ifaceName, enumName, doc.Synopsis(e.Description.Summary))
	fmt.Fprint(w, comment(e.Description.Text))
	fmt.Fprintf(w, "const (\n")
	for _, entry := range e.Entries {
		entryName := toCamel(entry.Name)

		if entry.Summary != "" {
			fmt.Fprintf(w, "// %s%s%s : %s\n", ifaceName, enumName, entryName, doc.Synopsis(entry.Summary))
		}
		fmt.Fprintf(w, "%s%s%s = %s\n", ifaceName, enumName, entryName, entry.Value)
	}
	fmt.Fprintf(w, ")\n")
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

				if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
					argIface = "client." + argIface
				}
				if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
					argIface = "xdg_shell." + argIface
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
	fmt.Fprintf(w, "type %s%sHandler interface {\n", ifaceName, eventName)
	fmt.Fprintf(w, "Handle%s%s(%s%sEvent)\n", ifaceName, eventName, ifaceName, eventName)
	fmt.Fprintf(w, "}\n")

	// Add handler
	fmt.Fprintf(w, "// Add%sHandler : adds handler for %s%sEvent\n", eventName, ifaceName, eventName)
	fmt.Fprintf(w, "func (i *%s) Add%sHandler(h %s%sHandler) {\n", ifaceName, eventName, ifaceName, eventName)
	fmt.Fprintf(w, "if h == nil {\n")
	fmt.Fprintf(w, "return\n")
	fmt.Fprintf(w, "}\n\n")
	fmt.Fprintf(w, "i.mu.Lock()\n")
	fmt.Fprintf(w, "i.%sHandlers = append(i.%sHandlers, h)\n", eventNameLower, eventNameLower)
	fmt.Fprintf(w, "i.mu.Unlock()\n")
	fmt.Fprintf(w, "}\n")

	// Remove handler
	fmt.Fprintf(w, "func (i *%s) Remove%sHandler(h %s%sHandler) {\n", ifaceName, eventName, ifaceName, eventName)
	fmt.Fprintf(w, "i.mu.Lock()\n")
	fmt.Fprintf(w, "defer i.mu.Unlock()\n\n")
	fmt.Fprintf(w, "for j, e := range i.%sHandlers {\n", eventNameLower)
	fmt.Fprintf(w, "if e == h {\n")
	fmt.Fprintf(w, "i.%sHandlers = append(i.%sHandlers[:j], i.%sHandlers[j+1:]...)\n", eventNameLower, eventNameLower, eventNameLower)
	fmt.Fprintf(w, "break\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "}\n")
}

func writeEventDispatcher(w io.Writer, ifaceName string, v Interface) {
	if len(v.Events) == 0 {
		return
	}

	if protocol.Name != "wayland" {
		fmt.Fprintf(w, "func (i *%s) Dispatch(event *client.Event) {\n", ifaceName)
	} else {
		fmt.Fprintf(w, "func (i *%s) Dispatch(event *Event) {\n", ifaceName)
	}
	fmt.Fprintf(w, "switch event.Opcode {\n")
	for i, e := range v.Events {
		eventName := toCamel(e.Name)
		eventNameLower := toLowerCamel(e.Name)

		fmt.Fprintf(w, "case %d:\n", i)
		fmt.Fprintf(w, "i.mu.RLock()\n")
		fmt.Fprintf(w, "if len(i.%sHandlers) == 0 {\n", eventNameLower)
		fmt.Fprintf(w, "i.mu.RUnlock()\n")
		fmt.Fprintf(w, "break\n")
		fmt.Fprintf(w, "}\n")
		fmt.Fprintf(w, "i.mu.RUnlock()\n\n")
		fmt.Fprintf(w, "e := %s%sEvent{\n", ifaceName, eventName)
		for _, arg := range e.Args {
			switch arg.Type {
			case "object", "new_id":
				if arg.Interface != "" {
					argName := toCamel(arg.Name)
					argIface := toCamel(arg.Interface)

					if protocol.Name != "wayland" && strings.HasPrefix(arg.Interface, "wl_") {
						argIface = "client." + argIface
					}
					if protocol.Name != "xdg_shell" && strings.HasPrefix(arg.Interface, "xdg_") {
						argIface = "xdg_shell." + argIface
					}

					fmt.Fprintf(w, "%s: event.Proxy(i.Context()).(*%s),\n", argName, argIface)
				} else {
					fmt.Fprintf(w, "%s: event.Proxy(i.Context()),\n", toCamel(arg.Name))
				}

			case "int", "uint", "fixed",
				"string", "array", "fd":
				fmt.Fprintf(w, "%s: event.%s(),\n", toCamel(arg.Name), typeToGetterMap[arg.Type])
			}
		}
		fmt.Fprintf(w, "}\n\n")

		fmt.Fprintf(w, "i.mu.RLock()\n")
		fmt.Fprintf(w, "for _, h := range i.%sHandlers {\n", eventNameLower)
		fmt.Fprintf(w, "i.mu.RUnlock()\n\n")
		fmt.Fprintf(w, "h.Handle%s%s(e)\n\n", ifaceName, eventName)
		fmt.Fprintf(w, "i.mu.RLock()\n")
		fmt.Fprintf(w, "}\n")
		fmt.Fprintf(w, "i.mu.RUnlock()\n")
	}
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "}\n")
}

func toCamel(s string) string {
	return strings.ReplaceAll(strcase.ToCamel(s), "Id", "ID")
}

func toLowerCamel(s string) string {
	return strings.ReplaceAll(strcase.ToLowerCamel(s), "Id", "ID")
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
