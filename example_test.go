package kongplete

func ExampleComplete() {
	var cli struct {
		Foo struct {
		} `kong:"cmd"`
	}

	_ = cli
}
