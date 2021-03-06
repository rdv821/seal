/*
Copyright © 2020 Infoblox <dev@infoblox.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/infobloxopen/seal/pkg/compiler"
	// register the rego backend compiler
	_ "github.com/infobloxopen/seal/pkg/compiler/rego"
	"github.com/spf13/cobra"
)

var compileSettings struct {
	files       []string // files to compile
	backend     string   // backend compiler
	outputFile  string   // output filename
	swaggerFile string   // swagger file to read in types
}

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Takes a list of seal inputs and compiles them",
	Long: `compile takes a list of seal inputs
and compiles them to a target authorization runtime.
You can use one of the built in runtimes or call
a custom backend to target your own runtime.`,
	Run: compileFunc,
}

func compileFunc(cmd *cobra.Command, args []string) {

	if compileSettings.swaggerFile == "" {
		log.Println("swagger file is required for inferring types")
		os.Exit(1)
	}

	swaggerSpec, err := ioutil.ReadFile(compileSettings.swaggerFile)
	if err != nil {
		log.Printf("could not read swagger file %v: %s", compileSettings.swaggerFile, err)
		os.Exit(1)
	}

	cplr, err := compiler.NewPolicyCompiler(compileSettings.backend, string(swaggerSpec))
	if err != nil {
		log.Printf("could not create policy compiler: %s", err)
		os.Exit(1)
	}

	var output []string
	for _, fil := range compileSettings.files {

		// read file
		input, err := ioutil.ReadFile(fil)
		if err != nil {
			log.Printf("could not read file %v: %s\n", fil, err)
			os.Exit(1)
		}

		// compile policies from policy rules
		pkgname := path.Base(fil)
		out, err := cplr.Compile(pkgname, string(input))
		if err != nil {
			log.Printf("could not compile file %v: %s\n", fil, err)
			os.Exit(1)
		}

		output = append(output, out)
	}

	// write to output
	switch compileSettings.outputFile {
	case "-", "":
		fmt.Printf("%s\n", strings.Join(output, "\n"))
	default:
		err := ioutil.WriteFile(compileSettings.outputFile,
			[]byte(fmt.Sprintf("%s\n", strings.Join(output, "\n"))),
			0644)
		if err != nil {
			log.Printf("could not write to file %v: %s\n", compileSettings.outputFile, err)
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.AddCommand(compileCmd)

	// Here you will define your flags and configuration settings.
	compileCmd.PersistentFlags().StringArrayVarP(&compileSettings.files, "file", "f", []string{},
		"filename or directory to read seal files")
	compileCmd.PersistentFlags().StringVarP(&compileSettings.backend, "backend", "b", "rego",
		"compiler backend")
	compileCmd.PersistentFlags().StringVarP(&compileSettings.swaggerFile, "swagger-file", "s", "",
		"filename to read types")
	compileCmd.PersistentFlags().StringVarP(&compileSettings.outputFile, "output", "o", "",
		"output file")
}
