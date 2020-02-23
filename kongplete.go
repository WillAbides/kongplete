package kongplete

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/posener/complete/cmd/install"
)

const predictorTag = "predictor"

type options struct {
	predictors   map[string]complete.Predictor
	exitFunc     func(code int)
	errorHandler func(error)
}

//Option is a configuration option for running Complete
type Option func(*options)

//WithPredictor use the named predictor
func WithPredictor(name string, predictor complete.Predictor) Option {
	return func(o *options) {
		if o.predictors == nil {
			o.predictors = map[string]complete.Predictor{}
		}
		o.predictors[name] = predictor
	}
}

//WithPredictors use these predictors
func WithPredictors(predictors map[string]complete.Predictor) Option {
	return func(o *options) {
		for k, v := range predictors {
			WithPredictor(k, v)(o)
		}
	}
}

//WithExitFunc the exit command that is run after completions
func WithExitFunc(exitFunc func(code int)) Option {
	return func(o *options) {
		o.exitFunc = exitFunc
	}
}

//WithErrorHandler handle errors with completions
func WithErrorHandler(handler func(error)) Option {
	return func(o *options) {
		o.errorHandler = handler
	}
}

func buildOptions(opt ...Option) *options {
	opts := &options{
		predictors: map[string]complete.Predictor{},
	}
	for _, o := range opt {
		o(opts)
	}
	return opts
}

//Command returns a completion Command for a kong parser
func Command(parser *kong.Kong, opt ...Option) (complete.Command, error) {
	opts := buildOptions(opt...)
	if parser == nil || parser.Model == nil {
		return complete.Command{}, nil
	}
	command, err := nodeCommand(parser.Model.Node, opts.predictors)
	if err != nil {
		return complete.Command{}, err
	}
	return *command, err
}

//Complete runs completion for a kong parser
func Complete(parser *kong.Kong, opt ...Option) {
	if parser == nil {
		return
	}
	opts := buildOptions(opt...)
	errHandler := opts.errorHandler
	if errHandler == nil {
		errHandler = func(err error) {
			parser.Errorf("error running command completion: %v", err)
		}
	}
	exitFunc := opts.exitFunc
	if exitFunc == nil {
		exitFunc = parser.Exit
	}
	cmd, err := Command(parser, opt...)
	if err != nil {
		errHandler(err)
		exitFunc(1)
	}
	cmp := complete.New(parser.Model.Name, cmd)
	cmp.Out = parser.Stdout
	done := cmp.Complete()
	if done {
		exitFunc(0)
	}
}

func nodeCommand(node *kong.Node, predictors map[string]complete.Predictor) (*complete.Command, error) {
	if node == nil {
		return nil, nil
	}

	cmd := complete.Command{
		Sub:         complete.Commands{},
		GlobalFlags: complete.Flags{},
	}

	for _, child := range node.Children {
		if child == nil {
			continue
		}
		childCmd, err := nodeCommand(child, predictors)
		if err != nil {
			return nil, err
		}
		if childCmd != nil {
			cmd.Sub[child.Name] = *childCmd
		}
	}

	for _, flag := range node.Flags {
		if flag == nil {
			continue
		}
		predictor, err := flagPredictor(flag, predictors)
		if err != nil {
			return nil, err
		}
		cmd.GlobalFlags["--"+flag.Name] = predictor
		if flag.Short != 0 {
			cmd.GlobalFlags["-"+string(flag.Short)] = predictor
		}
	}

	var err error
	cmd.Args, err = argsPredictor(node.Positional, predictors)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

//kongTag interface for *kong.kongTag
type kongTag interface {
	Has(string) bool
	Get(string) string
}

func tagPredictor(tag kongTag, predictors map[string]complete.Predictor) (complete.Predictor, error) {
	if tag == nil {
		return nil, nil
	}
	if !tag.Has(predictorTag) {
		return nil, nil
	}
	if predictors == nil {
		predictors = map[string]complete.Predictor{}
	}
	predictorName := tag.Get(predictorTag)
	predictor, ok := predictors[predictorName]
	if !ok {
		return nil, fmt.Errorf("no predictor with name %q", predictorName)
	}
	return predictor, nil
}

func valuePredictor(value *kong.Value, predictors map[string]complete.Predictor) (complete.Predictor, error) {
	if value == nil {
		return nil, nil
	}
	predictor, err := tagPredictor(value.Tag, predictors)
	if err != nil {
		return nil, err
	}
	if predictor != nil {
		return predictor, nil
	}
	switch {
	case value.IsBool():
		return complete.PredictNothing, nil
	case value.Enum != "":
		enumVals := make([]string, 0, len(value.EnumMap()))
		for enumVal := range value.EnumMap() {
			enumVals = append(enumVals, enumVal)
		}
		return complete.PredictSet(enumVals...), nil
	default:
		return complete.PredictAnything, nil
	}
}

func argsPredictor(args []*kong.Positional, predictors map[string]complete.Predictor) (complete.Predictor, error) {
	switch len(args) {
	case 0:
		return nil, nil
	case 1:
		return valuePredictor(args[0], predictors)
	}
	resPredictors := make([]complete.Predictor, 0, len(args))
	for _, arg := range args {
		resPredictor, err := valuePredictor(arg, predictors)
		if err != nil {
			return nil, err
		}
		resPredictors = append(resPredictors, resPredictor)
	}
	return complete.PredictOr(resPredictors...), nil
}

func flagPredictor(flag *kong.Flag, predictors map[string]complete.Predictor) (complete.Predictor, error) {
	return valuePredictor(flag.Value, predictors)
}

//InstallShellCompletions is a helper to install completions for a kong context
func InstallShellCompletions(k *kong.Context) error {
	return install.Install(k.Model.Name)
}

//UninstallShellCompletions is a helper to uninstall completions for a kong context
func UninstallShellCompletions(k *kong.Context) error {
	return install.Uninstall(k.Model.Name)
}

//InstallCompletions is a kong command for installing or uninstalling shell completions
type InstallCompletions struct {
	Uninstall bool
}

//Run runs InstallCompletions
func (c *InstallCompletions) Run(k *kong.Context) error {
	if c.Uninstall {
		return UninstallShellCompletions(k)
	}
	return InstallShellCompletions(k)
}
