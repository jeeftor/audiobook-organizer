package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var layoutTemplateCmd = &cobra.Command{
	Use:   "layout-template",
	Short: "Show custom layout template field reference",
	Long: `Display detailed information about custom organization layout templates.

Use this command to see available fields, fallback syntax, examples, and
directory path safety rules for the --layout-template flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprint(cmd.OutOrStdout(), layoutTemplateHelpText())
	},
}

func layoutTemplateHelpText() string {
	return `Audiobook Layout Template Field Reference

USAGE
  audiobook-organizer --dir=/books --out=/organized \
    --layout-template="{author}/{series|Standalone}/{Vol series_number:02 - }{title}{ [narrator]}"

PLACEHOLDER SYNTAX
  {field}              Render a metadata field
  ${field}             Alternate field syntax
  {field|Fallback}     Use fallback text when the field is missing or empty
  {field:02}           Zero-pad numeric fields such as series_number
  {Vol series_number:02 - }  Composite optional segment omitted when any inner field is empty

COMMON FIELDS
  {author}             First author, formatted with the organizer author format
  {authors}            All authors, comma-separated
  {title}              Book title
  {series}             Series name without number
  {series_full}        Series name with number when available
  {series_number}      Series number only, such as 1 or 2.5
  {series-count}       Alias for {series_number}
  {album}              Album field from audio metadata
  {track}              Track number, zero-padded to two digits
  {year}               Publication year from raw metadata
  {narrator}           First narrator or narrator value when available
  {narrators}          All narrators, comma-separated when available

RAW METADATA FIELDS
  Templates can also reference raw metadata keys. Dashes are normalized to
  underscores, so {publisher-name} can read a raw field named publisher_name.

EXAMPLES
  Author / Series / Vol NN - Title with optional narrator brackets
    {author}/{series|Standalone}/{Vol series_number:02 - }{title}{ [narrator]}

  Standalone fallback for books without a series
    {author}/{series|Standalone}/{title}

  Include narrator in the book folder. (Placeholders get omitted when empty)
    {author}/{series|Standalone}/{series-count} - {title} {(narrator)}    
  
  Omit empty slash-separated segments automatically
    {author}/{series}/{title}

PATH SAFETY
  Slashes in the template create directories. Metadata values are sanitized
  inside each path segment, so slashes or unsafe characters in metadata cannot
  create extra directories.

  Absolute templates and "." or ".." path segments are rejected.

MORE DOCS
  https://github.com/jeeftor/audiobook-organizer/blob/master/docs/LAYOUTS.md
`
}

func init() {
	rootCmd.AddCommand(layoutTemplateCmd)
}
