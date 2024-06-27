package link

type Config struct {
	Id        string
	Domain    string
	Host      string
	Target    string
	Remote    string
	StorePath string

	WindowSize     int
	RecoverSupport bool
}
