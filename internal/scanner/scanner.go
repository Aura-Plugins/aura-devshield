package scanner

type Scanner interface {
	Name() string
	Scan() ([]Finding, error)
}
