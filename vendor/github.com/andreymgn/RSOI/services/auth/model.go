package auth

type Auth interface {
	Add(string, string) (string, error)
	Exists(string) (bool, error)
}
