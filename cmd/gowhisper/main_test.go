package main

import (
	"testing"

	// Packages
	kong "github.com/alecthomas/kong"
)

// The command line interface builds, with no duplicate flags, and the
// transcribe and translate commands take the language of the audio
func Test_CLI(t *testing.T) {
	for _, args := range [][]string{
		{"transcribe", "model", "file", "--language", "de"},
		{"translate", "model", "file", "--language", "de"},
	} {
		cli := new(CLI)
		parser, err := kong.New(cli, kong.Vars{"version": VersionJSON(), "EXECUTABLE_NAME": "gowhisper"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parser.Parse(args); err != nil {
			t.Fatal(args, err)
		}
		var language *string
		switch args[0] {
		case "transcribe":
			language = cli.Transcribe.Language
		case "translate":
			language = cli.Translate.Language
		}
		if language == nil || *language != "de" {
			t.Errorf("%v: language %v", args, language)
		}
	}
}
