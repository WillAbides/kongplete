package kongplete

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	envLine  = "COMP_LINE"
	envPoint = "COMP_POINT"
)

func TestComplete(t *testing.T) {
	type embed struct {
		Lion string
	}

	predictors := map[string]complete.Predictor{
		"things":      complete.PredictSet("thing1", "thing2"),
		"otherthings": complete.PredictSet("otherthing1", "otherthing2"),
	}

	var cli struct {
		Foo struct {
			Embedded embed  `kong:"embed"`
			Bar      string `kong:"predictor=things"`
			Baz      bool
			Rabbit   struct {
			} `kong:"cmd"`
			Duck struct {
			} `kong:"cmd"`
		} `kong:"cmd"`
		Bar struct {
			Tiger  string `kong:"arg,predictor=things"`
			Bear   string `kong:"arg,predictor=otherthings"`
			OMG    string `kong:"enum='oh,my,gizzles'"`
			Number int    `kong:"short=n,enum='1,2,3'"`
		} `kong:"cmd"`
	}

	tests := []completeTest{
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"foo", "bar"},
			want:            true,
			line:            "myApp ",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"foo"},
			want:            true,
			line:            "myApp foo",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"rabbit", "duck"},
			want:            true,
			line:            "myApp foo ",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"rabbit"},
			want:            true,
			line:            "myApp foo r",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"--bar", "--baz", "--lion"},
			want:            true,
			line:            "myApp foo -",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{},
			want:            true,
			line:            "myApp foo --lion ",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"rabbit", "duck"},
			want:            true,
			line:            "myApp foo --baz ",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"--bar", "--baz", "--lion"},
			want:            true,
			line:            "myApp foo --baz -",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"thing1", "thing2"},
			want:            true,
			line:            "myApp foo --bar ",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"thing1", "thing2", "otherthing1", "otherthing2"},
			want:            true,
			line:            "myApp bar ",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"oh", "my", "gizzles"},
			want:            true,
			line:            "myApp bar --omg ",
			point:           -1,
		},
		{
			parser:          kong.Must(&cli),
			appName:         "myApp",
			wantCompletions: []string{"-n", "--number", "--omg"},
			want:            true,
			line:            "myApp bar -",
			point:           -1,
		},
	}

	for _, td := range tests {
		name := td.name
		if name == "" {
			name = td.line
		}
		t.Run(name, func(t *testing.T) {
			options := td.options
			if options == nil {
				options = []Option{WithPredictors(predictors)}
			}
			completions, got, err := runComplete(t, td.appName, td.parser, td.line, td.point, options)
			if td.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, td.want, got)
			assert.ElementsMatch(t, td.wantCompletions, completions)
		})
	}
}

func Test_tagPredictor(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got, err := tagPredictor(nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("no predictor tag", func(t *testing.T) {
		got, err := tagPredictor(testTag{}, nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("missing predictor", func(t *testing.T) {
		got, err := tagPredictor(testTag{predictorTag: "foo"}, nil)
		assert.Error(t, err)
		assert.Equal(t, `no predictor with name "foo"`, err.Error())
		assert.Nil(t, got)
	})

	t.Run("existing predictor", func(t *testing.T) {
		got, err := tagPredictor(testTag{predictorTag: "foo"}, map[string]complete.Predictor{"foo": complete.PredictAnything})
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
}

type testTag map[string]string

func (t testTag) Has(k string) bool {
	_, ok := t[k]
	return ok
}

func (t testTag) Get(k string) string {
	return t[k]
}

type completeTest struct {
	name            string
	parser          *kong.Kong
	appName         string
	options         []Option
	wantCompletions []string
	want            bool
	wantErr         bool
	line            string
	point           int
}

func setLineAndPoint(t *testing.T, line string, point int) func() {
	t.Helper()
	if point == -1 {
		point = len(line)
	}
	origLine, hasOrigLine := os.LookupEnv(envLine)
	origPoint, hasOrigPoint := os.LookupEnv(envPoint)
	require.NoError(t, os.Setenv(envLine, line))
	require.NoError(t, os.Setenv(envPoint, strconv.Itoa(point)))
	return func() {
		t.Helper()
		require.NoError(t, os.Unsetenv(envLine))
		require.NoError(t, os.Unsetenv(envPoint))
		if hasOrigLine {
			require.NoError(t, os.Setenv(envLine, origLine))
		}
		if hasOrigPoint {
			require.NoError(t, os.Setenv(envPoint, origPoint))
		}
	}
}

func runComplete(t *testing.T, appName string, parser *kong.Kong, line string, point int, opts []Option) ([]string, bool, error) {
	t.Helper()
	cleanup := setLineAndPoint(t, line, point)
	defer cleanup()
	var buf bytes.Buffer
	if parser != nil {
		parser.Stdout = &buf
	}
	got, err := Complete(appName, parser, opts...)
	return parseOutput(buf.String()), got, err
}

func parseOutput(output string) []string {
	lines := strings.Split(output, "\n")
	options := []string{}
	for _, l := range lines {
		if l != "" {
			options = append(options, l)
		}
	}
	return options
}
