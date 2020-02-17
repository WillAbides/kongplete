package kongplete

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
)

const predictorTag = "predictor"

type options struct {
	predictors map[string]complete.Predictor
}

type Option func(*options)

func WithPredictor(name string, predictor complete.Predictor) Option {
	return func(o *options) {
		if o.predictors == nil {
			o.predictors = map[string]complete.Predictor{}
		}
		o.predictors[name] = predictor
	}
}

func WithPredictors(predictors map[string]complete.Predictor) Option {
	return func(o *options) {
		for k, v := range predictors {
			WithPredictor(k, v)(o)
		}
	}
}

//Command returns a completion Command for a kong parser
func Command(parser *kong.Kong, opt ...Option) (complete.Command, error) {
	opts := &options{}
	for _, o := range opt {
		o(opts)
	}
	if parser == nil || parser.Model == nil {
		return complete.Command{}, nil
	}
	command, err := nodeCommand(parser.Model.Node, opts.predictors)
	if err != nil {
		return complete.Command{}, err
	}
	return *command, err
}

//Complete runs completion for a kong parser and returns true if completions ran (you usually want to exit when it returns true)
func Complete(appName string, parser *kong.Kong, opt ...Option) (bool, error) {
	if parser == nil {
		return false, nil
	}
	cmd, err := Command(parser, opt...)
	if err != nil {
		return false, err
	}
	cmp := complete.New(appName, cmd)
	cmp.Out = parser.Stdout
	return cmp.Complete(), nil
}

func nodeCommand(node *kong.Node, predictors map[string]complete.Predictor) (*complete.Command, error) {
	if node == nil {
		return nil, nil
	}

	cmd := complete.Command{
		Sub:   complete.Commands{},
		Flags: complete.Flags{},
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
		cmd.Flags["--"+flag.Name] = predictor
		if flag.Short != 0 {
			cmd.Flags["-"+string(flag.Short)] = predictor
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
