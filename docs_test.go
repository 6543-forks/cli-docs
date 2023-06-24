package docs

import (
	"bytes"
	"io"
	"io/fs"
	"net/mail"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func expectFileContent(t *testing.T, file, got string) {
	data, err := os.ReadFile(file)
	// Ignore windows line endings
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))

	r := require.New(t)
	r.NoError(err)
	r.Equal(got, string(data))
}

func buildExtendedTestCommand() *cli.Command {
	return &cli.Command{
		Writer: io.Discard,
		Name:   "greet",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "socket",
				Aliases:   []string{"s"},
				Usage:     "some 'usage' text",
				Value:     "value",
				TakesFile: true,
			},
			&cli.StringFlag{Name: "flag", Aliases: []string{"fl", "f"}},
			&cli.BoolFlag{
				Name:    "another-flag",
				Aliases: []string{"b"},
				Usage:   "another usage text",
				Sources: cli.EnvVars("EXAMPLE_VARIABLE_NAME"),
			},
			&cli.BoolFlag{
				Name:   "hidden-flag",
				Hidden: true,
			},
		},
		Commands: []*cli.Command{{
			Aliases: []string{"c"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:      "flag",
					Aliases:   []string{"fl", "f"},
					TakesFile: true,
				},
				&cli.BoolFlag{
					Name:    "another-flag",
					Aliases: []string{"b"},
					Usage:   "another usage text",
				},
			},
			Name:  "config",
			Usage: "another usage test",
			Commands: []*cli.Command{{
				Aliases: []string{"s", "ss"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "sub-flag", Aliases: []string{"sub-fl", "s"}},
					&cli.BoolFlag{
						Name:    "sub-command-flag",
						Aliases: []string{"s"},
						Usage:   "some usage text",
					},
				},
				Name:  "sub-config",
				Usage: "another usage test",
			}},
		}, {
			Aliases: []string{"i", "in"},
			Name:    "info",
			Usage:   "retrieve generic information",
		}, {
			Name: "some-command",
		}, {
			Name:   "hidden-command",
			Hidden: true,
		}, {
			Aliases: []string{"u"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:      "flag",
					Aliases:   []string{"fl", "f"},
					TakesFile: true,
				},
				&cli.BoolFlag{
					Name:    "another-flag",
					Aliases: []string{"b"},
					Usage:   "another usage text",
				},
			},
			Name:  "usage",
			Usage: "standard usage text",
			UsageText: `
Usage for the usage text
- formatted:  Based on the specified ConfigMap and summon secrets.yml
- list:       Inspect the environment for a specific process running on a Pod
- for_effect: Compare 'namespace' environment with 'local'

` + "```" + `
func() { ... }
` + "```" + `

Should be a part of the same code block
`,
			Commands: []*cli.Command{{
				Aliases: []string{"su"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "sub-command-flag",
						Aliases: []string{"s"},
						Usage:   "some usage text",
					},
				},
				Name:      "sub-usage",
				Usage:     "standard usage text",
				UsageText: "Single line of UsageText",
			}},
		}},
		UsageText:   "app [first_arg] [second_arg]",
		Description: `Description of the application.`,
		Usage:       "Some app",
		Authors: []any{
			"Harrison <harrison@lolwut.example.com>",
			&mail.Address{Name: "Oliver Allen", Address: "oliver@toyshop.com"},
		},
	}
}

func TestToMarkdownFull(t *testing.T) {
	// Given
	cmd := buildExtendedTestCommand()

	// When
	res, err := ToMarkdown(cmd)

	// Then
	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-doc-full.md", res)
}

func TestToTabularMarkdown(t *testing.T) {
	app := buildExtendedTestCommand()

	t.Run("full", func(t *testing.T) {
		// When
		res, err := ToTabularMarkdown(app, "app")

		// Then
		require.NoError(t, err)
		expectFileContent(t, "testdata/expected-tabular-markdown-full.md", res)
	})

	t.Run("with empty path", func(t *testing.T) {
		// When
		res, err := ToTabularMarkdown(app, "")

		// Then
		require.NoError(t, err)
		expectFileContent(t, "testdata/expected-tabular-markdown-full.md", res)
	})

	t.Run("with custom app path", func(t *testing.T) {
		// When
		res, err := ToTabularMarkdown(app, "/usr/local/bin")

		// Then
		require.NoError(t, err)
		expectFileContent(t, "testdata/expected-tabular-markdown-custom-app-path.md", res)
	})
}

func TestToTabularMarkdownFailed(t *testing.T) {
	tpl := MarkdownTabularDocTemplate
	t.Cleanup(func() { MarkdownTabularDocTemplate = tpl })

	MarkdownTabularDocTemplate = "{{ .Foo }}" // invalid template

	app := buildExtendedTestCommand()

	res, err := ToTabularMarkdown(app, "")

	r := require.New(t)
	r.Error(err)

	r.Equal("", res)
}

func TestToTabularToFileBetweenTags(t *testing.T) {
	expectedDocs, fErr := os.ReadFile("testdata/expected-tabular-markdown-full.md")

	r := require.New(t)
	r.NoError(fErr)

	// normalizes \r\n (windows) and \r (mac) into \n (unix) (required for tests to pass on windows)
	normalizeNewlines := func(d []byte) []byte {
		d = bytes.ReplaceAll(d, []byte{13, 10}, []byte{10}) // replace CR LF \r\n (windows) with LF \n (unix)
		return bytes.ReplaceAll(d, []byte{13}, []byte{10})  // replace CF \r (mac) with LF \n (unix)
	}

	t.Run("default tags", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "")

		r := require.New(t)
		r.NoError(err)

		t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })

		_, err = tmpFile.WriteString(`# App readme file

Some description

<!--GENERATED:CLI_DOCS-->
<!--/GENERATED:CLI_DOCS-->

Some other text`)
		r.NoError(err)
		_ = tmpFile.Close()

		r.NoError(ToTabularToFileBetweenTags(buildExtendedTestCommand(), "app", tmpFile.Name()))

		content, err := os.ReadFile(tmpFile.Name())
		r.NoError(err)

		content = normalizeNewlines(content)

		expected := normalizeNewlines([]byte(`# App readme file

Some description

<!--GENERATED:CLI_DOCS-->
<!-- Documentation inside this block generated by github.com/urfave/cli; DO NOT EDIT -->
` + string(expectedDocs) + `
<!--/GENERATED:CLI_DOCS-->

Some other text`))

		r.Equal(string(expected), string(content))
	})

	t.Run("custom tags", func(t *testing.T) {
		r := require.New(t)

		tmpFile, err := os.CreateTemp("", "")
		r.NoError(err)

		t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) }) // cleanup

		_, err = tmpFile.WriteString(`# App readme file

Some description

foo_BAR|baz
lorem+ipsum

Some other text`)
		r.NoError(err)
		_ = tmpFile.Close()

		r.NoError(ToTabularToFileBetweenTags(buildExtendedTestCommand(), "app", tmpFile.Name(), "foo_BAR|baz", "lorem+ipsum"))

		content, err := os.ReadFile(tmpFile.Name())
		r.NoError(err)

		content = normalizeNewlines(content)

		expected := normalizeNewlines([]byte(`# App readme file

Some description

foo_BAR|baz
<!-- Documentation inside this block generated by github.com/urfave/cli; DO NOT EDIT -->
` + string(expectedDocs) + `
lorem+ipsum

Some other text`))

		r.Equal(string(expected), string(content))
	})

	t.Run("missing file", func(t *testing.T) {
		r := require.New(t)

		tmpFile, err := os.CreateTemp("", "")
		r.NoError(err)
		_ = tmpFile.Close()

		r.NoError(os.Remove(tmpFile.Name()))

		err = ToTabularToFileBetweenTags(buildExtendedTestCommand(), "app", tmpFile.Name())

		r.ErrorIs(err, fs.ErrNotExist)
	})
}

func TestToMarkdownNoFlags(t *testing.T) {
	app := buildExtendedTestCommand()
	app.Flags = nil

	res, err := ToMarkdown(app)

	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-doc-no-flags.md", res)
}

func TestToMarkdownNoCommands(t *testing.T) {
	app := buildExtendedTestCommand()
	app.Commands = nil

	res, err := ToMarkdown(app)

	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-doc-no-commands.md", res)
}

func TestToMarkdownNoAuthors(t *testing.T) {
	app := buildExtendedTestCommand()
	app.Authors = []any{}

	res, err := ToMarkdown(app)

	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-doc-no-authors.md", res)
}

func TestToMarkdownNoUsageText(t *testing.T) {
	app := buildExtendedTestCommand()
	app.UsageText = ""

	res, err := ToMarkdown(app)

	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-doc-no-usagetext.md", res)
}

func TestToMan(t *testing.T) {
	app := buildExtendedTestCommand()

	res, err := ToMan(app)

	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-doc-full.man", res)
}

func TestToManParseError(t *testing.T) {
	app := buildExtendedTestCommand()

	tmp := MarkdownDocTemplate
	t.Cleanup(func() { MarkdownDocTemplate = tmp })

	MarkdownDocTemplate = "{{ .App.Name"
	_, err := ToMan(app)

	require.ErrorContains(t, err, "template: cli:1: unclosed action")
}

func TestToManWithSection(t *testing.T) {
	cmd := buildExtendedTestCommand()

	res, err := ToManWithSection(cmd, 8)

	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-doc-full.man", res)
}

func Test_prepareUsageText(t *testing.T) {
	t.Run("no UsageText provided", func(t *testing.T) {
		cmd := &cli.Command{}
		res := prepareUsageText(cmd)
		require.Equal(t, "", res)
	})

	t.Run("single line UsageText", func(t *testing.T) {
		cmd := &cli.Command{UsageText: "Single line usage text"}
		res := prepareUsageText(cmd)
		require.Equal(t, ">Single line usage text\n", res)
	})

	t.Run("multiline UsageText", func(t *testing.T) {
		cmd := &cli.Command{
			UsageText: `
Usage for the usage text
- Should be a part of the same code block
`,
		}

		res := prepareUsageText(cmd)

		require.Equal(t, `    Usage for the usage text
    - Should be a part of the same code block
`, res)
	})

	t.Run("multiline UsageText has formatted embedded markdown", func(t *testing.T) {
		cmd := &cli.Command{
			UsageText: `
Usage for the usage text

` + "```" + `
func() { ... }
` + "```" + `

Should be a part of the same code block
`,
		}

		res := prepareUsageText(cmd)

		require.Equal(t, `    Usage for the usage text
    
    `+"```"+`
    func() { ... }
    `+"```"+`
    
    Should be a part of the same code block
`, res)
	})
}

func Test_prepareUsage(t *testing.T) {
	t.Run("no Usage provided", func(t *testing.T) {
		cmd := &cli.Command{}
		res := prepareUsage(cmd, "")
		require.Equal(t, "", res)
	})

	t.Run("simple Usage", func(t *testing.T) {
		cmd := &cli.Command{Usage: "simple usage text"}
		res := prepareUsage(cmd, "")
		require.Equal(t, cmd.Usage+"\n", res)
	})

	t.Run("simple Usage with UsageText", func(t *testing.T) {
		cmd := &cli.Command{Usage: "simple usage text"}
		res := prepareUsage(cmd, "a non-empty string")
		require.Equal(t, cmd.Usage+"\n\n", res)
	})
}
