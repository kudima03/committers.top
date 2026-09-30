package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"most-active-github-users-counter/merge"
	"most-active-github-users-counter/output"
	"most-active-github-users-counter/top"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var locations arrayFlags
var excludeLocations arrayFlags
var presetTitle string
var presetChecksum string

func main() {
	token := flag.String("token", LookupEnvOrString("GITHUB_TOKEN", ""), "Github auth token")
	amount := flag.Int("amount", 256, "Amount of users to show")
	considerNum := flag.Int("consider", 1000, "Amount of users to consider")
	outputOpt := flag.String("output", "plain", "Output format: plain, csv")
	fileName := flag.String("file", "", "Output file (optional, defaults to stdout)")
	presetName := flag.String("preset", "", "Preset (optional)")
	listPresets := flag.Bool("list-presets", false, "List all available presets as CSV and exit immediately")
	mergeDir := flag.String("merge", "", "Merge regional YAML files from this directory into a global rating (optional)")

	flag.Var(&locations, "location", "Location to query")
	flag.Parse()

	if *listPresets {
		fmt.Println("preset,title,definition_checksum")
		for name, _ := range PRESETS {
			fmt.Printf("%v,\"%v\",%v\n", name, PresetTitle(name), PresetChecksum(name))
		}
		return
	}

	if *presetName != "" {
		preset := Preset(*presetName)
		locations = preset.include
		excludeLocations = preset.exclude
		presetTitle = PresetTitle(*presetName)
		presetChecksum = PresetChecksum(*presetName)
	}

	if *mergeDir != "" {
		if *outputOpt != "yaml" {
			log.Fatal("Merging supports only yaml output")
		}
		mergeAmount := *amount
		if mergeAmount > merge.MaxAmount {
			log.Printf("Amount %v is too large for merging, using %v", mergeAmount, merge.MaxAmount)
			mergeAmount = merge.MaxAmount
		}
		locations, err := merge.LoadDir(*mergeDir, merge.Slug)
		if err != nil {
			log.Fatal(err)
		}
		writeOutput(*fileName, func(writer *bufio.Writer) error {
			return output.GlobalYamlOutput(merge.Merge(locations, mergeAmount), writer)
		})
		return
	}

	var format output.Format

	if *outputOpt == "plain" {
		format = output.PlainOutput
	} else if *outputOpt == "yaml" {
		format = output.YamlOutput
	} else if *outputOpt == "csv" {
		format = output.CsvOutput
	} else {
		log.Fatal("Unrecognized output format: ", *outputOpt)
	}

	opts := top.Options{Token: *token, Locations: locations, ExcludeLocations: excludeLocations, Amount: *amount, ConsiderNum: *considerNum, PresetTitle: presetTitle, PresetChecksum: presetChecksum}
	data, err := top.GithubTop(opts)

	if err != nil {
		log.Fatal(err)
	}

	writeOutput(*fileName, func(writer *bufio.Writer) error {
		return format(data, writer, opts)
	})
}

func writeOutput(fileName string, write func(*bufio.Writer) error) {
	var writer *bufio.Writer
	if fileName != "" {
		f, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(f)
		defer f.Close()
	} else {
		writer = bufio.NewWriter(os.Stdout)
	}

	if err := write(writer); err != nil {
		log.Fatal(err)
	}
	writer.Flush()
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
