package discovery

type Display struct {
	ID        string
	Name      string
	Primary   bool
	X         int
	Y         int
	Width     int
	Height    int
	Connector string
}

type Window struct {
	ID     string
	Title  string
	X      int
	Y      int
	Width  int
	Height int
}

type Camera struct {
	ID       string
	Label    string
	Device   string
	CardName string
}

type AudioInput struct {
	ID         string
	Name       string
	Driver     string
	SampleSpec string
	State      string
}

type Snapshot struct {
	Displays []Display
	Windows  []Window
	Cameras  []Camera
	Audio    []AudioInput
}
