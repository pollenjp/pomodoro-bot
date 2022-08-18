package app

type destructor struct {
	destructor []func()
}

func (d *destructor) Append(f func()) {
	d.destructor = append(d.destructor, f)
}

func (d *destructor) Close() {
	for _, f := range d.destructor {
		f()
	}
}

var (
	Destructor = &destructor{}
)
