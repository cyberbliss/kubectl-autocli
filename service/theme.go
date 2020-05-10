package service

import "github.com/c-bata/go-prompt"

// prompt.Color is just an int - so use values between 1 and 255
// to find out what the colours look like on your terminal execute
// _example/print256colours.sh
// Any option with a value of 0 will use the terminal's default colour

type Theme struct {
	OptionDescriptionBGColor prompt.Color
	OptionDescriptionTextColor prompt.Color
	OptionInputBGColor prompt.Color
	OptionInputTextColor prompt.Color
	OptionPrefixBackgroundColor prompt.Color
	OptionPrefixTextColor prompt.Color
	OptionPreviewSuggestionBGColor prompt.Color
	OptionPreviewSuggestionTextColor prompt.Color
	OptionScrollbarBGColor prompt.Color
	OptionScrollbarThumbColor prompt.Color
	OptionSelectedDescriptionBGColor prompt.Color
	OptionSelectedDescriptionTextColor prompt.Color
	OptionSelectedSuggestionBGColor prompt.Color
	OptionSelectedSuggestionTextColor prompt.Color
	OptionSuggestionBGColor prompt.Color
	OptionSuggestionTextColor prompt.Color
}

var Themes = map[string]Theme {
	"light": {
		OptionDescriptionBGColor:           159,
		OptionDescriptionTextColor:         239,
		OptionInputBGColor:                 0,
		OptionInputTextColor:               40,
		OptionPrefixBackgroundColor:        0,
		OptionPrefixTextColor:              208,
		OptionPreviewSuggestionBGColor:     0,
		OptionPreviewSuggestionTextColor:   40,
		OptionScrollbarBGColor:             243,
		OptionScrollbarThumbColor:          253,
		OptionSelectedDescriptionBGColor:   18,
		OptionSelectedDescriptionTextColor: 220,
		OptionSelectedSuggestionBGColor:    18,
		OptionSelectedSuggestionTextColor:  220,
		OptionSuggestionBGColor:            117,
		OptionSuggestionTextColor:          18,
	},
}
