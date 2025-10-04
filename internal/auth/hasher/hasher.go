package hasher

type HasherInterface interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}
